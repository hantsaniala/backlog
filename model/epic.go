package model

type Epic struct {
	ID       string   `yaml:"id"`
	Name     string   `yaml:"name"`
	Status   Status   `yaml:"status"`
	Priority Priority `yaml:"priority"`
	Labels   []string `yaml:"labels"`
	DueDate  string   `yaml:"due_date"`
	Created  string   `yaml:"created"`
	Updated  string   `yaml:"updated"`

	Body      string  `yaml:"-"`
	Filename  string  `yaml:"-"`
	ProjectID string  `yaml:"-"`
	Children  []*Task `yaml:"-"`
}

func (e *Epic) FullID() string {
	return e.ID
}

func (e *Epic) StoryCount() int {
	count := 0
	for _, c := range e.Children {
		if c.Type == TypeStory {
			count++
		}
	}
	return count
}
