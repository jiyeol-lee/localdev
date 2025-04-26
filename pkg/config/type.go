package config

type ConfigCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
	Silent      bool   `json:"silent"`
	AutoExecute bool   `json:"autoExecute"`
}

type ConfigCommands struct {
	LowerA *ConfigCommand `json:"lowerA,omitempty"`
	LowerB *ConfigCommand `json:"lowerB,omitempty"`
	LowerC *ConfigCommand `json:"lowerC,omitempty"`
	LowerD *ConfigCommand `json:"lowerD,omitempty"`
	LowerE *ConfigCommand `json:"lowerE,omitempty"`
	LowerF *ConfigCommand `json:"lowerF,omitempty"`
	LowerG *ConfigCommand `json:"lowerG,omitempty"`
	LowerH *ConfigCommand `json:"lowerH,omitempty"`
	LowerI *ConfigCommand `json:"lowerI,omitempty"`
	LowerJ *ConfigCommand `json:"lowerJ,omitempty"`
	LowerK *ConfigCommand `json:"lowerK,omitempty"`
	LowerL *ConfigCommand `json:"lowerL,omitempty"`
	LowerM *ConfigCommand `json:"lowerM,omitempty"`
	LowerN *ConfigCommand `json:"lowerN,omitempty"`
	LowerO *ConfigCommand `json:"lowerO,omitempty"`
	LowerP *ConfigCommand `json:"lowerP,omitempty"`
	LowerQ *ConfigCommand `json:"lowerQ,omitempty"`
	LowerR *ConfigCommand `json:"lowerR,omitempty"`
	LowerS *ConfigCommand `json:"lowerS,omitempty"`
	LowerT *ConfigCommand `json:"lowerT,omitempty"`
	LowerU *ConfigCommand `json:"lowerU,omitempty"`
	LowerV *ConfigCommand `json:"lowerV,omitempty"`
	LowerW *ConfigCommand `json:"lowerW,omitempty"`
	LowerX *ConfigCommand `json:"lowerX,omitempty"`
	LowerY *ConfigCommand `json:"lowerY,omitempty"`
	LowerZ *ConfigCommand `json:"lowerZ,omitempty"`

	UpperA *ConfigCommand `json:"upperA,omitempty"`
	UpperB *ConfigCommand `json:"upperB,omitempty"`
	UpperC *ConfigCommand `json:"upperC,omitempty"`
	UpperD *ConfigCommand `json:"upperD,omitempty"`
	UpperE *ConfigCommand `json:"upperE,omitempty"`
	UpperF *ConfigCommand `json:"upperF,omitempty"`
	UpperG *ConfigCommand `json:"upperG,omitempty"`
	UpperH *ConfigCommand `json:"upperH,omitempty"`
	UpperI *ConfigCommand `json:"upperI,omitempty"`
	UpperJ *ConfigCommand `json:"upperJ,omitempty"`
	UpperK *ConfigCommand `json:"upperK,omitempty"`
	UpperL *ConfigCommand `json:"upperL,omitempty"`
	UpperM *ConfigCommand `json:"upperM,omitempty"`
	UpperN *ConfigCommand `json:"upperN,omitempty"`
	UpperO *ConfigCommand `json:"upperO,omitempty"`
	UpperP *ConfigCommand `json:"upperP,omitempty"`
	UpperQ *ConfigCommand `json:"upperQ,omitempty"`
	UpperR *ConfigCommand `json:"upperR,omitempty"`
	UpperS *ConfigCommand `json:"upperS,omitempty"`
	UpperT *ConfigCommand `json:"upperT,omitempty"`
	UpperU *ConfigCommand `json:"upperU,omitempty"`
	UpperV *ConfigCommand `json:"upperV,omitempty"`
	UpperW *ConfigCommand `json:"upperW,omitempty"`
	UpperX *ConfigCommand `json:"upperX,omitempty"`
	UpperY *ConfigCommand `json:"upperY,omitempty"`
	UpperZ *ConfigCommand `json:"upperZ,omitempty"`
}

type ConfigPane struct {
	Name     string          `json:"name"`
	Dir      string          `json:"dir"`
	Start    string          `json:"start"`
	Stop     string          `json:"stop"`
	Commands *ConfigCommands `json:"commands,omitempty"`
}

type Config struct {
	Panes []ConfigPane `json:"panes"`
}
