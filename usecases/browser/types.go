package browser

type actionType string

const (
	ActionQuestion    actionType = "question"
	ActionPageRequest actionType = "page_request"
	ActionNavigate    actionType = "navigate"
	ActionClick       actionType = "click"
	ActionFill        actionType = "fill"
	ActionExtract     actionType = "extract"
	ActionScroll      actionType = "scroll"
	ActionWait        actionType = "wait"
	ActionScreenshot  actionType = "screenshot"
)

type alertActionType string

func (a alertActionType) String() string {
	return string(a)
}

const (
	AlertAccept  alertActionType = "accept"
	AlertDismiss alertActionType = "dismiss"
)

type scrollDirection string

const (
	ScrollUp    scrollDirection = "up"
	ScrollDown  scrollDirection = "down"
	ScrollLeft  scrollDirection = "left"
	ScrollRight scrollDirection = "right"
)

type stepsResponse struct {
	Done  bool   `json:"done,omitempty"`
	Steps []step `json:"steps,omitempty"`
}

type step struct {
	Action actionType `json:"actionType"`

	Question   *questionStep   `json:"question,omitempty"`
	Navigate   *navigateStep   `json:"navigate,omitempty"`
	Click      *clickStep      `json:"click,omitempty"`
	Fill       *fillStep       `json:"fill,omitempty"`
	Extract    *extractStep    `json:"extract,omitempty"`
	Scroll     *scrollStep     `json:"scroll,omitempty"`
	Wait       *waitStep       `json:"wait,omitempty"`
	Screenshot *screenshotStep `json:"screenshot,omitempty"`
}

type questionStep struct {
	Question string `json:"question"`
}

type navigateStep struct {
	URL string `json:"url"`
}

type clickStep struct {
	Selector string `json:"selector"`
}

type fillStep struct {
	Selector  string `json:"selector"`
	FillValue string `json:"fill_value"`
}

type extractStep struct {
	Selector string `json:"extract_selector"`
}

type scrollStep struct {
	Direction scrollDirection `json:"direction"`
	Amount    int             `json:"amount"`
}

type waitStep struct {
	Selector string `json:"wait_selector"`
	Timeout  int    `json:"wait_timeout"`
}

type screenshotStep struct {
	Path string `json:"screenshot_path"`
}
