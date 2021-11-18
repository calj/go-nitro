package protocols

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/statechannels/go-nitro/channel/state"
	"github.com/statechannels/go-nitro/types"
)

func TestNew(t *testing.T) {
	_, err := NewDirectFundingObjectiveState(state.TestState, state.TestState.Participants[0])
	if err != nil {
		t.Error(err)
	}
}

var s, _ = NewDirectFundingObjectiveState(state.TestState, state.TestState.Participants[0])
var dummySignature = state.Signature{
	R: common.Hex2Bytes(`49d8e91bd182fb4d489bb2d76a6735d494d5bea24e4b51dd95c9d219293312d9`),
	S: common.Hex2Bytes(`22274a3cec23c31e0c073b3c071cf6e0c21260b0d292a10e6a04257a2d8e87fa`),
	V: byte(1),
}
var dummyStateHash = common.Hash{}
var stateToSign state.State = s.ExpectedStates[0]
var stateHash, _ = stateToSign.Hash()
var correctSignatureByParticipant, _ = stateToSign.Sign(common.Hex2Bytes(`caab404f975b4620747174a75f08d98b4e5a7053b691b41bcfc0d839d48b7634`))

func TestUpdate(t *testing.T) {
	// First, prepare a new objective using the constructor:
	s, _ := NewDirectFundingObjectiveState(state.TestState, state.TestState.Participants[0])
	// Next, prepare an event with a mismatched channelId
	e := ObjectiveEvent{
		ChannelId: types.Destination{},
	}
	// Assert that this should return an error
	// TODO is this the behaviour we want? Below with the signatures, we prefer a log + NOOP (no error)
	_, err := s.Update(e)
	if err == nil {
		t.Error(`ChannelId mismatch -- expected an error but did not get one`)
	}

	// Now modify the event to give it the "correct" channelId (matching the objective),
	// and make a new Sigs map.
	// This prepares us for the rest of the test.
	e.ChannelId = s.ChannelId
	e.Sigs = make(map[types.Bytes32]state.Signature)

	// Next, attempt to update the objective with a dummy signature, keyed with a dummy statehash
	// Assert that this results in a NOOP
	e.Sigs[dummyStateHash] = dummySignature // Dummmy signature on dummy statehash
	_, err = s.Update(e)
	if err != nil {
		t.Error(`dummy signature -- expected a noop but caught an error:`, err)
	}

	// Next, attempt to update the objective with an invalid signature, keyed with a dummy statehash
	// Assert that this results in a NOOP
	e.Sigs[dummyStateHash] = state.Signature{}
	_, err = s.Update(e)
	if err != nil {
		t.Error(`faulty signature -- expected a noop but caught an error:`, err)
	}

	// Next, attempt to update the objective with correct signature by a participant on a relevant state
	// Assert that this results in an appropriate change in the extended state of the objective
	e.Sigs[stateHash] = correctSignatureByParticipant
	updated, err := s.Update(e)
	if err != nil {
		t.Error(err)
	}
	if updated.(DirectFundingObjectiveState).PreFundSigned[0] != true {
		t.Error(`Objective data not updated as expected`)
	}

}
