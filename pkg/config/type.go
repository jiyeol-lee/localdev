package config

type ConfigCommand struct {
	Command     string `yaml:"command"`
	Description string `yaml:"description"`
	Silent      bool   `yaml:"silent"`
	AutoExecute bool   `yaml:"autoExecute"`
}

type ConfigCommands struct {
	LowerA *ConfigCommand `yaml:"lowerA,omitempty"`
	LowerB *ConfigCommand `yaml:"lowerB,omitempty"`
	LowerC *ConfigCommand `yaml:"lowerC,omitempty"`
	LowerD *ConfigCommand `yaml:"lowerD,omitempty"`
	LowerE *ConfigCommand `yaml:"lowerE,omitempty"`
	LowerF *ConfigCommand `yaml:"lowerF,omitempty"`
	LowerG *ConfigCommand `yaml:"lowerG,omitempty"`
	LowerH *ConfigCommand `yaml:"lowerH,omitempty"`
	LowerI *ConfigCommand `yaml:"lowerI,omitempty"`
	LowerJ *ConfigCommand `yaml:"lowerJ,omitempty"`
	LowerK *ConfigCommand `yaml:"lowerK,omitempty"`
	LowerL *ConfigCommand `yaml:"lowerL,omitempty"`
	LowerM *ConfigCommand `yaml:"lowerM,omitempty"`
	LowerN *ConfigCommand `yaml:"lowerN,omitempty"`
	LowerO *ConfigCommand `yaml:"lowerO,omitempty"`
	LowerP *ConfigCommand `yaml:"lowerP,omitempty"`
	LowerQ *ConfigCommand `yaml:"lowerQ,omitempty"`
	LowerR *ConfigCommand `yaml:"lowerR,omitempty"`
	LowerS *ConfigCommand `yaml:"lowerS,omitempty"`
	LowerT *ConfigCommand `yaml:"lowerT,omitempty"`
	LowerU *ConfigCommand `yaml:"lowerU,omitempty"`
	LowerV *ConfigCommand `yaml:"lowerV,omitempty"`
	LowerW *ConfigCommand `yaml:"lowerW,omitempty"`
	LowerX *ConfigCommand `yaml:"lowerX,omitempty"`
	LowerY *ConfigCommand `yaml:"lowerY,omitempty"`
	LowerZ *ConfigCommand `yaml:"lowerZ,omitempty"`

	UpperA *ConfigCommand `yaml:"upperA,omitempty"`
	UpperB *ConfigCommand `yaml:"upperB,omitempty"`
	UpperC *ConfigCommand `yaml:"upperC,omitempty"`
	UpperD *ConfigCommand `yaml:"upperD,omitempty"`
	UpperE *ConfigCommand `yaml:"upperE,omitempty"`
	UpperF *ConfigCommand `yaml:"upperF,omitempty"`
	UpperG *ConfigCommand `yaml:"upperG,omitempty"`
	UpperH *ConfigCommand `yaml:"upperH,omitempty"`
	UpperI *ConfigCommand `yaml:"upperI,omitempty"`
	UpperJ *ConfigCommand `yaml:"upperJ,omitempty"`
	UpperK *ConfigCommand `yaml:"upperK,omitempty"`
	UpperL *ConfigCommand `yaml:"upperL,omitempty"`
	UpperM *ConfigCommand `yaml:"upperM,omitempty"`
	UpperN *ConfigCommand `yaml:"upperN,omitempty"`
	UpperO *ConfigCommand `yaml:"upperO,omitempty"`
	UpperP *ConfigCommand `yaml:"upperP,omitempty"`
	UpperQ *ConfigCommand `yaml:"upperQ,omitempty"`
	UpperR *ConfigCommand `yaml:"upperR,omitempty"`
	UpperS *ConfigCommand `yaml:"upperS,omitempty"`
	UpperT *ConfigCommand `yaml:"upperT,omitempty"`
	UpperU *ConfigCommand `yaml:"upperU,omitempty"`
	UpperV *ConfigCommand `yaml:"upperV,omitempty"`
	UpperW *ConfigCommand `yaml:"upperW,omitempty"`
	UpperX *ConfigCommand `yaml:"upperX,omitempty"`
	UpperY *ConfigCommand `yaml:"upperY,omitempty"`
	UpperZ *ConfigCommand `yaml:"upperZ,omitempty"`
}

type ProjectSettings struct {
	Dir     string `yaml:"dir,omitempty"`
	Command string `yaml:"command,omitempty"`
}

type ConfigPane struct {
	Name     string          `yaml:"name"`
	Dir      string          `yaml:"dir"`
	Start    string          `yaml:"start"`
	Stop     string          `yaml:"stop"`
	Commands *ConfigCommands `yaml:"commands,omitempty"`
}

type Config struct {
	ProjectSettings *ProjectSettings `yaml:"project_settings,omitempty"`
	Panes           []ConfigPane     `yaml:"panes"`
}
