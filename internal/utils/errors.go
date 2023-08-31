package utils

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	sdkresource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/labd/apollostudio-go-sdk/apollostudio"
)

func NewRetryableError(err error) *sdkresource.RetryError {
	if err == nil {
		return nil
	}
	if apollostudio.IsOperationError(err) {
		return sdkresource.NonRetryableError(err)
	}
	return sdkresource.RetryableError(err)
}

func ProcessError(diags *diag.Diagnostics, err error, w, e string) {
	if err != nil {
		if apollostudio.IsOperationError(err) {
			diags.AddWarning(w, err.Error())
			return
		}
		diags.AddError(
			e,
			err.Error(),
		)
	}
}
