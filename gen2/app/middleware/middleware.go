package middleware

import (
	"github.com/ebi-yade/cloud-functions-samples/gen2/app"
)

// FIXME: 戻り値を http.HandlerFunc にする
func WrapForHTTP(mids []app.MiddlewareForHTTP, handler app.HTTPHandlerFunc) app.HTTPHandlerFunc {
	for i := len(mids) - 1; i >= 0; i-- {
		mwFunc := mids[i] // 実質 pop
		if mwFunc != nil {
			handler = mwFunc(handler)
		}
	}

	return handler
}
