package model

type Sprint struct {
	ID      string `yaml:"id"`
	Name    string `yaml:"name"`
	Start   string `yaml:"start"`
	End     string `yaml:"end"`
	Goal    string `yaml:"goal"`
	Created string `yaml:"created"`

	Filename  string `yaml:"-"`
	ProjectID string `yaml:"-"`
	Tasks     []*Task `yaml:"-"`
}

func (s *Sprint) TotalPoints() int {
	total := 0
	for _, t := range s.Tasks {
		if t.StoryPoints != nil {
			total += *t.StoryPoints
		}
	}
	return total
}

func (s *Sprint) CompletedPoints() int {
	total := 0
	for _, t := range s.Tasks {
		if t.StoryPoints != nil && t.Status == StatusDone {
			total += *t.StoryPoints
		}
	}
	return total
}
