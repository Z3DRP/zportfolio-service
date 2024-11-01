package adapters

type TaskData struct {
	FmtTaskInfo string
	Details     string
	Method      string
}

type UserData struct {
	Name    string
	Company string
	Email   string
	Phone   string
	Roles   []string
}

type Customizations struct {
	BodyColor   string
	TextColor   string
	BannerImage string // use base 64 encode image
}

type TaskRequestAlertDTO struct {
	TaskData
	UserData
	Customizations
}

type ThanksAlertDTO struct {
	UserData
	Customizations
}
