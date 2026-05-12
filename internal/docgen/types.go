package docgen

type RoadmapNode struct {
	Name      string
	DependsOn []string
}

type EvidenceLink struct {
	Role     string
	Feature  string
	Step     string
	Handoff  string
	Outputs  []string
	EdgeCase int
}

type FeatureState struct {
	Feature string
	Step    string
	Status  string
}

type Data struct {
	Title        string
	Project      string
	RoadmapText  string
	FeatureDocs  []string
	PlanDocs     []string
	Specs        []string
	Scenarios    int
	RoadmapNodes []RoadmapNode
	Evidence     []EvidenceLink
	States       []FeatureState
	KB           []KBPage
}

type KBPage struct {
	Feature      string
	Summary      string
	Audience     string
	Status       string
	WhatItDoes   string
	WhenToUse    string
	HowItBehaves string
	Examples     string
}
