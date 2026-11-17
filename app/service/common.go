package service

import (
	"errors"

	"hunispace/app/repository"
	"hunispace/helper"
)

func lookupError(err error, what string) error {
	if errors.Is(err, repository.ErrNotFound) {
		return helper.NotFound(what + " tidak ditemukan")
	}
	return helper.Internal(err)
}
