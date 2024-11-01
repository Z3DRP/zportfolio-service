package services

type Service interface {
	Initialize() Service
	Execute() error
}
