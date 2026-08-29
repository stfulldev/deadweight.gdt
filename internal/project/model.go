package project

// Request contains all inputs required to discover a Godot project root.
type Request struct {
	SceneInput       string
	WorkingDirectory string
	ExplicitProject  string
}

// Root identifies a discovered Godot project and its marker file.
type Root struct {
	Directory   string
	ProjectFile string
}
