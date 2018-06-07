package models

type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type LabelSnapshotConfig struct {
	Label                *Label `json:"label"`
	IntervalSeconds      int64  `json:"intervalSeconds"`
	RetentionPeriodHours int64  `json:"retentionPeriodHours"`
}

type Description struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type DescriptionSnapshotConfig struct {
	Description          *Description `json:"description"`
	IntervalSeconds      int64        `json:"intervalSeconds"`
	RetentionPeriodHours int64        `json:"retentionPeriodHours"`
}

type SnapshotConfigs struct {
	Descriptions []*DescriptionSnapshotConfig `json:"Descriptions"`
	Labels       []*LabelSnapshotConfig       `json:"Labels"`
}
