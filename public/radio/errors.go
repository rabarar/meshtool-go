package radio

import "errors"

var ErrUnkownPayloadType = errors.New("unknown payload type")
var ErrDecrypt = errors.New("unable to decrypt payload")
