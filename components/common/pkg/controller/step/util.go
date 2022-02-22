package step

import (
	apierrs "k8s.io/apimachinery/pkg/api/errors"
)

func isHardFailure(err error) {
	switch {
	case apierrs.IsBadRequest(err), apierrs.IsInvalid(err):

	}
}
