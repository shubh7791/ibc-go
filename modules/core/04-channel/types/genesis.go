package types

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"

	host "github.com/cosmos/ibc-go/v10/modules/core/24-host"
	ibcerrors "github.com/cosmos/ibc-go/v10/modules/core/errors"
)

// NewPacketState creates a new PacketState instance.
func NewPacketState(portID, channelID string, seq uint64, data []byte) PacketState {
	return PacketState{
		PortId:    portID,
		ChannelId: channelID,
		Sequence:  seq,
		Data:      data,
	}
}

// Validate performs basic validation of fields returning an error upon any
// failure.
func (pa PacketState) Validate() error {
	if pa.Data == nil {
		return errors.New("data bytes cannot be nil")
	}
	return validateGenFields(pa.PortId, pa.ChannelId, pa.Sequence)
}

// NewPacketSequence creates a new PacketSequences instance.
func NewPacketSequence(portID, channelID string, seq uint64) PacketSequence {
	return PacketSequence{
		PortId:    portID,
		ChannelId: channelID,
		Sequence:  seq,
	}
}

// Validate performs basic validation of fields returning an error upon any
// failure.
func (ps PacketSequence) Validate() error {
	return validateGenFields(ps.PortId, ps.ChannelId, ps.Sequence)
}

// NewGenesisState creates a GenesisState instance.
func NewGenesisState(
	channels []IdentifiedChannel, acks, receipts, commitments []PacketState,
	sendSeqs, recvSeqs, ackSeqs []PacketSequence, nextChannelSequence uint64,
	params Params,
) GenesisState {
	return GenesisState{
		Channels:            channels,
		Acknowledgements:    acks,
		Receipts:            receipts,
		Commitments:         commitments,
		SendSequences:       sendSeqs,
		RecvSequences:       recvSeqs,
		AckSequences:        ackSeqs,
		NextChannelSequence: nextChannelSequence,
		Params:              params,
	}
}

// DefaultGenesisState returns the ibc channel submodule's default genesis state.
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Channels:            []IdentifiedChannel{},
		Acknowledgements:    []PacketState{},
		Receipts:            []PacketState{},
		Commitments:         []PacketState{},
		SendSequences:       []PacketSequence{},
		RecvSequences:       []PacketSequence{},
		AckSequences:        []PacketSequence{},
		NextChannelSequence: 0,
		Params:              DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// keep track of the max sequence to ensure it is less than
	// the next sequence used in creating connection identifiers.
	var maxSequence uint64

	for i, channel := range gs.Channels {
		sequence, err := ParseChannelSequence(channel.ChannelId)
		if err != nil {
			return err
		}

		if sequence > maxSequence {
			maxSequence = sequence
		}

		if err := channel.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid channel %v channel index %d: %w", channel, i, err)
		}
	}

	if maxSequence != 0 && maxSequence >= gs.NextChannelSequence {
		return errorsmod.Wrapf(ibcerrors.ErrInvalidSequence, "next channel sequence %d must be greater than maximum sequence used in channel identifier %d", gs.NextChannelSequence, maxSequence)
	}

	for i, ack := range gs.Acknowledgements {
		if err := ack.Validate(); err != nil {
			return errorsmod.Wrapf(ErrInvalidAcknowledgement, "%v ack index %d: %s", ack, i, err.Error())
		}
		if len(ack.Data) == 0 {
			return errorsmod.Wrapf(ErrInvalidAcknowledgement, "%v ack index %d: data bytes cannot be empty", ack, i)
		}
	}

	for i, receipt := range gs.Receipts {
		if err := receipt.Validate(); err != nil {
			return errorsmod.Wrapf(ErrInvalidAcknowledgement, "%v ack index %d: %s", receipt, i, err.Error())
		}
	}

	for i, commitment := range gs.Commitments {
		if err := commitment.Validate(); err != nil {
			return errorsmod.Wrapf(ErrInvalidCommitment, "%v index %d: %s", commitment, i, err.Error())
		}
		if len(commitment.Data) == 0 {
			return errorsmod.Wrapf(ErrInvalidAcknowledgement, "%v ack index %d: data bytes cannot be empty", commitment, i)
		}
	}

	for i, ss := range gs.SendSequences {
		if err := ss.Validate(); err != nil {
			return errorsmod.Wrapf(ibcerrors.ErrInvalidSequence, "invalid send sequence %v index %d: %s", ss, i, err.Error())
		}
	}

	for i, rs := range gs.RecvSequences {
		if err := rs.Validate(); err != nil {
			return fmt.Errorf("invalid receive sequence %v index %d: %w", rs, i, err)
		}
	}

	for i, as := range gs.AckSequences {
		if err := as.Validate(); err != nil {
			return errorsmod.Wrapf(ErrInvalidAcknowledgement, "sequence %v index %d: %s", as, i, err.Error())
		}
	}

	return nil
}

func validateGenFields(portID, channelID string, sequence uint64) error {
	if err := host.PortIdentifierValidator(portID); err != nil {
		return fmt.Errorf("invalid port Id: %w", err)
	}
	if err := host.ChannelIdentifierValidator(channelID); err != nil {
		return fmt.Errorf("invalid channel Id: %w", err)
	}
	if sequence == 0 {
		return errors.New("sequence cannot be 0")
	}
	return nil
}
