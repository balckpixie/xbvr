package config

type DmmAPIKey struct {
	DmmApiId       string `json:"dmmApiId"`
	DmmAffiliateId string `json:"dmmAffiliateId"`
}

type ThumbnailSchedule struct {
	Enabled         bool `default:"false" json:"enabled"`
	HourInterval    int  `default:"2" json:"hourInterval"`
	UseRange        bool `default:"false" json:"useRange"`
	MinuteStart     int  `default:"0" json:"minuteStart"`
	HourStart       int  `default:"0" json:"hourStart"`
	HourEnd         int  `default:"23" json:"hourEnd"`
	RunAtStartDelay int  `default:"0" json:"runAtStartDelay"`
}

type ThumbnailParams struct {
	Start    int    `default:"15" json:"start"`
	Interval int    `default:"15" json:"interval"`
	Resolution    int    `default:"200" json:"resolution"`
	UseCUDAEncode	bool `default:"true" json:"useCUDAEncode"`
}
