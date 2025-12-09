package cmd

// GlobalFlags holds flags shared across all commands.
type GlobalFlags struct {
	LogLevel string
}

// PostArgs holds flags for the post command.
type PostArgs struct {
	*GlobalFlags

	Org             string
	Repo            string
	Token           string
	SHA1            string
	Template        string
	TemplateKey     string
	ConfigPath      string
	PRNumber        int
	Vars            []string
	VarFiles        []string
	DryRun          bool
	SkipNoToken     bool
	Silent          bool
	StdinTemplate   bool
	UpdateCondition string
}

// ExecArgs holds flags for the exec command.
type ExecArgs struct {
	*GlobalFlags

	Org         string
	Repo        string
	Token       string
	SHA1        string
	Template    string
	TemplateKey string
	ConfigPath  string
	PRNumber    int
	Outputs     []string
	Vars        []string
	VarFiles    []string
	DryRun      bool
	SkipNoToken bool
	Silent      bool
	Args        []string
}

// HideArgs holds flags for the hide command.
type HideArgs struct {
	*GlobalFlags

	Org         string
	Repo        string
	Token       string
	ConfigPath  string
	Condition   string
	HideKey     string
	PRNumber    int
	SHA1        string
	Vars        []string
	VarFiles    []string
	DryRun      bool
	SkipNoToken bool
	Silent      bool
}
