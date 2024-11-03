package services

import "context"

type Service interface {
	Execute(context.Context) error
}
