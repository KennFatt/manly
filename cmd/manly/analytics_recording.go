package main

import (
	"context"

	"github.com/KennFatt/manly/internal/analytics"
	"github.com/KennFatt/manly/internal/knowledge"
)

func recordConceptLoads(recorder analytics.Recorder, entryPoint analytics.EntryPoint, conceptIDs []string) {
	if recorder == nil || len(conceptIDs) == 0 {
		return
	}
	loads, err := analytics.NewBatch(entryPoint, conceptIDs)
	if err != nil {
		return
	}
	_ = recorder.RecordConceptLoads(context.Background(), loads)
}

func conceptIDs(concepts []*knowledge.Concept) []string {
	ids := make([]string, 0, len(concepts))
	for _, concept := range concepts {
		ids = append(ids, concept.ID)
	}
	return ids
}

func workspaceConceptIDs(workspace *knowledge.Workspace, refs []knowledge.ConceptRef) []string {
	ids := make([]string, 0, len(refs))
	for _, ref := range refs {
		ids = append(ids, workspace.QualifyID(ref.BundleName, ref.Concept.ID))
	}
	return ids
}
