package traceability

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

const (
	requirementsAccepted   = "accepted"
	requirementsProposed   = "proposed"
	requirementsSuperseded = "superseded"
	deliveryPlanned        = "planned"
	deliveryInProgress     = "in_progress"
	deliveryVerified       = "verified"
)

var (
	useCaseHeadingPattern = regexp.MustCompile(`(?m)^### Use Case ([0-9]+[A-Z]?):`)
	storyFilenamePattern  = regexp.MustCompile(`^us-(\d{3})-.*\.md$`)
	incrementHeading      = regexp.MustCompile(`(?m)^## (INC-\d{2})\b`)
	traceLinePattern      = regexp.MustCompile(`(?m)^\s*//\s*Trace:\s*(.+?)\s*$`)
	traceReferencePattern = regexp.MustCompile(`^(UC-[0-9]+[A-Z]?|US-\d{3})@r([1-9][0-9]*)$`)
	incrementReference    = regexp.MustCompile(`^INC-\d{2}$`)
	mergeSubjectPRPattern = regexp.MustCompile(`\(#([1-9][0-9]*)\)\s*$`)
)

type Manifest struct {
	SchemaVersion int                  `yaml:"schema_version"`
	Baseline      Baseline             `yaml:"baseline"`
	UseCases      map[string]UseCase   `yaml:"use_cases"`
	Stories       map[string]Story     `yaml:"stories"`
	Increments    map[string]Increment `yaml:"increments"`
}

type Baseline struct {
	RequirementsCommit      string `yaml:"requirements_commit"`
	UseCaseDocumentVersion  string `yaml:"use_case_document_version"`
	RequirementsDescription string `yaml:"requirements_description,omitempty"`
}

type UseCase struct {
	CurrentRevision int               `yaml:"current_revision"`
	Revisions       []UseCaseRevision `yaml:"revisions"`
}

type UseCaseRevision struct {
	Revision           int    `yaml:"revision"`
	RequirementsStatus string `yaml:"requirements_status"`
	Change             string `yaml:"change,omitempty"`
}

type Story struct {
	Path            string          `yaml:"path"`
	Increment       string          `yaml:"increment"`
	CurrentRevision int             `yaml:"current_revision"`
	Revisions       []StoryRevision `yaml:"revisions"`
}

type StoryRevision struct {
	Revision           int      `yaml:"revision"`
	RequirementsStatus string   `yaml:"requirements_status"`
	DeliveryStatus     string   `yaml:"delivery_status"`
	SourceUseCases     []string `yaml:"source_use_cases"`
	ImplementationPR   int      `yaml:"implementation_pr,omitempty"`
	Change             string   `yaml:"change,omitempty"`
}

type Increment struct {
	DeliveryStatus   string `yaml:"delivery_status"`
	ImplementationPR int    `yaml:"implementation_pr,omitempty"`
}

func Load(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode traceability manifest: %w", err)
	}
	return manifest, nil
}

func Validate(root string, manifest Manifest) []string {
	var problems []string
	add := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf(format, args...))
	}

	if manifest.SchemaVersion != 1 {
		add("schema_version must be 1")
	}
	if manifest.Baseline.RequirementsCommit == "" {
		add("baseline requirements_commit is required")
	}
	if manifest.Baseline.UseCaseDocumentVersion == "" {
		add("baseline use_case_document_version is required")
	}

	documentUseCases, err := useCasesInDocument(filepath.Join(root, "docs/use-cases.md"))
	if err != nil {
		add("%v", err)
	}
	documentStories, err := storiesInDocuments(filepath.Join(root, "docs/user-stories"))
	if err != nil {
		add("%v", err)
	}
	documentIncrements, err := incrementsInDocument(filepath.Join(root, "docs/user-stories/roadmap.md"))
	if err != nil {
		add("%v", err)
	}

	for _, id := range sortedKeys(documentUseCases) {
		if _, ok := manifest.UseCases[id]; !ok {
			add("%s exists in docs/use-cases.md but not the manifest", id)
		}
	}
	for _, id := range sortedKeys(manifest.UseCases) {
		if _, ok := documentUseCases[id]; !ok {
			add("%s exists in the manifest but not docs/use-cases.md", id)
		}
		validateUseCase(id, manifest.UseCases[id], add)
	}

	for _, id := range sortedKeys(documentStories) {
		if _, ok := manifest.Stories[id]; !ok {
			add("%s exists in docs/user-stories but not the manifest", id)
		}
	}
	for _, id := range sortedKeys(manifest.Stories) {
		story := manifest.Stories[id]
		path, ok := documentStories[id]
		if !ok {
			add("%s exists in the manifest but not docs/user-stories", id)
		} else if filepath.ToSlash(story.Path) != filepath.ToSlash(path) {
			add("%s path is %q, want %q", id, story.Path, path)
		}
		if _, ok := manifest.Increments[story.Increment]; !ok {
			add("%s references unknown increment %s", id, story.Increment)
		}
		validateStory(id, story, manifest, add)
	}

	for _, id := range sortedKeys(documentIncrements) {
		if _, ok := manifest.Increments[id]; !ok {
			add("%s exists in the roadmap but not the manifest", id)
		}
	}
	for _, id := range sortedKeys(manifest.Increments) {
		if _, ok := documentIncrements[id]; !ok {
			add("%s exists in the manifest but not the roadmap", id)
		}
		increment := manifest.Increments[id]
		if !validDeliveryStatus(increment.DeliveryStatus) {
			add("%s has invalid delivery status %q", id, increment.DeliveryStatus)
		}
		if increment.DeliveryStatus == deliveryVerified && id == "INC-01" && increment.ImplementationPR == 0 {
			add("INC-01 is verified but has no implementation_pr")
		}
	}

	traces, traceProblems := acceptanceTraces(filepath.Join(root, "acceptance/cases"), manifest)
	problems = append(problems, traceProblems...)
	for _, id := range sortedKeys(manifest.Stories) {
		story := manifest.Stories[id]
		current, ok := currentStoryRevision(story)
		if ok && current.DeliveryStatus == deliveryVerified {
			ref := fmt.Sprintf("%s@r%d", id, current.Revision)
			if !traces[ref] {
				add("%s is verified but has no matching acceptance trace", ref)
			}
		}
	}

	sort.Strings(problems)
	return problems
}

func validateUseCase(id string, useCase UseCase, add func(string, ...any)) {
	seen := make(map[int]bool)
	for _, revision := range useCase.Revisions {
		if revision.Revision < 1 {
			add("%s has revision less than 1", id)
		}
		if seen[revision.Revision] {
			add("%s has duplicate revision %d", id, revision.Revision)
		}
		seen[revision.Revision] = true
		if !validRequirementsStatus(revision.RequirementsStatus) {
			add("%s@r%d has invalid requirements status %q", id, revision.Revision, revision.RequirementsStatus)
		}
		if revision.Revision == useCase.CurrentRevision && revision.RequirementsStatus == requirementsSuperseded {
			add("%s@r%d is current but superseded", id, revision.Revision)
		}
		if revision.Revision != useCase.CurrentRevision && revision.RequirementsStatus != requirementsSuperseded {
			add("%s@r%d is historical but not superseded", id, revision.Revision)
		}
	}
	if !seen[useCase.CurrentRevision] {
		add("%s current revision %d does not exist", id, useCase.CurrentRevision)
	}
}

func validateStory(id string, story Story, manifest Manifest, add func(string, ...any)) {
	seen := make(map[int]bool)
	for _, revision := range story.Revisions {
		if revision.Revision < 1 {
			add("%s has revision less than 1", id)
		}
		if seen[revision.Revision] {
			add("%s has duplicate revision %d", id, revision.Revision)
		}
		seen[revision.Revision] = true
		if !validRequirementsStatus(revision.RequirementsStatus) {
			add("%s@r%d has invalid requirements status %q", id, revision.Revision, revision.RequirementsStatus)
		}
		if !validDeliveryStatus(revision.DeliveryStatus) {
			add("%s@r%d has invalid delivery status %q", id, revision.Revision, revision.DeliveryStatus)
		}
		if revision.Revision == story.CurrentRevision && revision.RequirementsStatus == requirementsSuperseded {
			add("%s@r%d is current but superseded", id, revision.Revision)
		}
		if revision.Revision != story.CurrentRevision && revision.RequirementsStatus != requirementsSuperseded {
			add("%s@r%d is historical but not superseded", id, revision.Revision)
		}
		if len(revision.SourceUseCases) == 0 {
			add("%s@r%d has no source_use_cases", id, revision.Revision)
		}
		for _, ref := range revision.SourceUseCases {
			parts := traceReferencePattern.FindStringSubmatch(ref)
			if parts == nil || !strings.HasPrefix(ref, "UC-") {
				add("%s@r%d has invalid source use-case reference %q", id, revision.Revision, ref)
				continue
			}
			source, ok := manifest.UseCases[parts[1]]
			sourceRevision, _ := strconv.Atoi(parts[2])
			if !ok || !hasUseCaseRevision(source, sourceRevision) {
				add("%s@r%d references unknown %s", id, revision.Revision, ref)
			}
		}
		if revision.DeliveryStatus == deliveryVerified && revision.ImplementationPR == 0 {
			add("%s@r%d is verified but has no implementation_pr", id, revision.Revision)
		}
	}
	if !seen[story.CurrentRevision] {
		add("%s current revision %d does not exist", id, story.CurrentRevision)
	}
}

func validRequirementsStatus(status string) bool {
	return status == requirementsProposed || status == requirementsAccepted || status == requirementsSuperseded
}

func validDeliveryStatus(status string) bool {
	return status == deliveryPlanned || status == deliveryInProgress || status == deliveryVerified
}

func useCasesInDocument(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read use cases: %w", err)
	}
	result := make(map[string]bool)
	for _, match := range useCaseHeadingPattern.FindAllStringSubmatch(string(data), -1) {
		result["UC-"+match[1]] = true
	}
	return result, nil
}

func storiesInDocuments(root string) (map[string]string, error) {
	result := make(map[string]string)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		match := storyFilenamePattern.FindStringSubmatch(entry.Name())
		if match == nil {
			return nil
		}
		repositoryPath, err := filepath.Rel(filepath.Dir(filepath.Dir(root)), path)
		if err != nil {
			return err
		}
		result["US-"+match[1]] = filepath.ToSlash(repositoryPath)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read user stories: %w", err)
	}
	return result, nil
}

func incrementsInDocument(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read roadmap: %w", err)
	}
	result := make(map[string]bool)
	for _, match := range incrementHeading.FindAllStringSubmatch(string(data), -1) {
		result[match[1]] = true
	}
	return result, nil
}

func acceptanceTraces(root string, manifest Manifest) (map[string]bool, []string) {
	traces := make(map[string]bool)
	var problems []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, line := range traceLinePattern.FindAllStringSubmatch(string(data), -1) {
			for _, ref := range strings.Fields(line[1]) {
				if incrementReference.MatchString(ref) {
					if _, ok := manifest.Increments[ref]; !ok {
						problems = append(problems, fmt.Sprintf("acceptance trace references unknown %s", ref))
					}
					traces[ref] = true
					continue
				}
				parts := traceReferencePattern.FindStringSubmatch(ref)
				if parts == nil {
					problems = append(problems, fmt.Sprintf("acceptance trace has invalid reference %q", ref))
					continue
				}
				revision, _ := strconv.Atoi(parts[2])
				if strings.HasPrefix(parts[1], "UC-") {
					item, ok := manifest.UseCases[parts[1]]
					if !ok || !hasUseCaseRevision(item, revision) {
						problems = append(problems, fmt.Sprintf("acceptance trace references unknown %s", ref))
					}
				} else {
					item, ok := manifest.Stories[parts[1]]
					if !ok || !hasStoryRevision(item, revision) {
						problems = append(problems, fmt.Sprintf("acceptance trace references unknown %s", ref))
					}
				}
				traces[ref] = true
			}
		}
		return nil
	})
	if err != nil {
		problems = append(problems, fmt.Sprintf("read acceptance traces: %v", err))
	}
	return traces, problems
}

func hasUseCaseRevision(useCase UseCase, revision int) bool {
	return slices.ContainsFunc(useCase.Revisions, func(candidate UseCaseRevision) bool {
		return candidate.Revision == revision
	})
}

func hasStoryRevision(story Story, revision int) bool {
	return slices.ContainsFunc(story.Revisions, func(candidate StoryRevision) bool {
		return candidate.Revision == revision
	})
}

func currentStoryRevision(story Story) (StoryRevision, bool) {
	for _, revision := range story.Revisions {
		if revision.Revision == story.CurrentRevision {
			return revision, true
		}
	}
	return StoryRevision{}, false
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func ParseMergeLog(log string) map[int]string {
	result := make(map[int]string)
	for line := range strings.Lines(log) {
		fields := strings.SplitN(strings.TrimSpace(line), "\t", 2)
		if len(fields) != 2 {
			continue
		}
		match := mergeSubjectPRPattern.FindStringSubmatch(fields[1])
		if match == nil {
			continue
		}
		number, err := strconv.Atoi(match[1])
		if err == nil {
			result[number] = fields[0]
		}
	}
	return result
}

func RenderReport(manifest Manifest, mergeSHAs map[int]string) string {
	var report strings.Builder
	report.WriteString("# Requirements traceability\n\n")
	report.WriteString("Generated from `traceability/manifest.yaml`. Do not edit this file directly.\n\n")
	fmt.Fprintf(&report, "Requirements baseline: use-case document version %s at `%s`.\n\n",
		manifest.Baseline.UseCaseDocumentVersion, manifest.Baseline.RequirementsCommit)

	report.WriteString("## Increments\n\n")
	report.WriteString("| Increment | Delivery status | Implementation PR | Merge SHA |\n")
	report.WriteString("|---|---|---:|---|\n")
	for _, id := range sortedKeys(manifest.Increments) {
		increment := manifest.Increments[id]
		pr, sha := evidence(increment.ImplementationPR, mergeSHAs)
		fmt.Fprintf(&report, "| %s | %s | %s | %s |\n", id, increment.DeliveryStatus, pr, sha)
	}

	report.WriteString("\n## Use cases and user stories\n\n")
	report.WriteString("| Use-case revision | Story revision | Increment | Requirements | Delivery | Implementation PR | Merge SHA |\n")
	report.WriteString("|---|---|---|---|---|---:|---|\n")
	for _, storyID := range sortedKeys(manifest.Stories) {
		story := manifest.Stories[storyID]
		revision, ok := currentStoryRevision(story)
		if !ok {
			continue
		}
		pr, sha := evidence(revision.ImplementationPR, mergeSHAs)
		for _, source := range revision.SourceUseCases {
			fmt.Fprintf(&report, "| %s | %s@r%d | %s | %s | %s | %s | %s |\n",
				source, storyID, revision.Revision, story.Increment, revision.RequirementsStatus,
				revision.DeliveryStatus, pr, sha)
		}
	}
	return report.String()
}

func evidence(prNumber int, mergeSHAs map[int]string) (string, string) {
	if prNumber == 0 {
		return "—", "—"
	}
	pr := fmt.Sprintf("#%d", prNumber)
	sha, ok := mergeSHAs[prNumber]
	if !ok {
		return pr, "pending merge"
	}
	return pr, "`" + sha + "`"
}
