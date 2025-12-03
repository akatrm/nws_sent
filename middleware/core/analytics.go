package core

type TrainingStatus struct {
	Running bool `json:"running"`
}

type TrainExamples struct {
	Label string `json:"label"`
	Text  string `json:"text"`
}

type Training struct {
	Examples []TrainExamples `json:"examples"`
}
