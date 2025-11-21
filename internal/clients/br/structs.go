package br

type Server struct {
	ID        int    `json:"id"`
	SymID     string `json:"sym_id"`
	Color     string `json:"color"`
	Name      string `json:"name"`
	MaxOnline int    `json:"max_online"`
	X2Enabled bool   `json:"x2"`
	Online    int    `json:"online"`
}

type Techwork struct {
	Enabled bool `json:"enabled"`
}

type Highlight struct {
	ID       int      `json:"id"`
	Title    string   `json:"title"`
	Media    Media    `json:"media"`
	Category Category `json:"category"`
}

type Media struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type News struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Media       Media  `json:"media"`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}
