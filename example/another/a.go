package another

func AnotherFunction() {

}

type EmailNotification struct {
}

func (s *EmailNotification) Send(msg string) bool {
	return true
}
