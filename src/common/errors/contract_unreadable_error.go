package errors

import stderrors "errors"

// ErrContractUnreadable marks a contract that answers with a revert or garbage, unlike an unreachable node.
var ErrContractUnreadable = stderrors.New("contract is unreadable")
