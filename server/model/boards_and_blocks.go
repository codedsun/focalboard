package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/mattermost/focalboard/server/utils"
	"github.com/mattermost/mattermost-server/v6/shared/mlog"
)

var NoBoardsInBoardsAndBlocksErr = errors.New("at least one board is required")
var NoBlocksInBoardsAndBlocksErr = errors.New("at least one block is required")
var NoTeamInBoardsAndBlocksErr = errors.New("team ID cannot be empty")
var BoardIDsAndPatchesMissmatchInBoardsAndBlocksErr = errors.New("board ids and patches need to match")
var BlockIDsAndPatchesMissmatchInBoardsAndBlocksErr = errors.New("block ids and patches need to match")

type BlockDoesntBelongToAnyBoardErr struct {
	blockID string
}

func (e BlockDoesntBelongToAnyBoardErr) Error() string {
	return fmt.Sprintf("block %s doesn't belong to any board", e.blockID)
}

// BoardsAndBlocks is used to operate over boards and blocks at the
// same time
// swagger:model
type BoardsAndBlocks struct {
	// The boards
	// required: false
	Boards []*Board `json:"boards"`

	// The blocks
	// required: false
	Blocks []Block `json:"blocks"`
}

func (bab *BoardsAndBlocks) IsValid() error {
	if len(bab.Boards) == 0 {
		return NoBoardsInBoardsAndBlocksErr
	}

	if len(bab.Blocks) == 0 {
		return NoBlocksInBoardsAndBlocksErr
	}

	boardsMap := map[string]bool{}
	for _, board := range bab.Boards {
		boardsMap[board.ID] = true
	}

	for _, block := range bab.Blocks {
		if _, ok := boardsMap[block.BoardID]; !ok {
			return BlockDoesntBelongToAnyBoardErr{block.ID}
		}
	}
	return nil
}

// DeleteBoardsAndBlocks is used to list the boards and blocks to
// delete on a request
// swagger:model
type DeleteBoardsAndBlocks struct {
	// The boards
	// required: true
	Boards []string `json:"boards"`

	// The blocks
	// required: true
	Blocks []string `json:"blocks"`
}

func (dbab *DeleteBoardsAndBlocks) IsValid() error {
	if len(dbab.Boards) == 0 {
		return NoBoardsInBoardsAndBlocksErr
	}

	if len(dbab.Blocks) == 0 {
		return NoBlocksInBoardsAndBlocksErr
	}

	return nil
}

// PatchBoardsAndBlocks is used to patch multiple boards and blocks on
// a single request
// swagger:model
type PatchBoardsAndBlocks struct {
	// The board IDs to patch
	// required: true
	BoardIDs []string `json:"boardIDs"`

	// The board patches
	// required: true
	BoardPatches []*BoardPatch `json:"boardPatches"`

	// The block IDs to patch
	// required: true
	BlockIDs []string `json:"blockIDs"`

	// The block patches
	// required: true
	BlockPatches []*BlockPatch `json:"blockPatches"`
}

func (dbab *PatchBoardsAndBlocks) IsValid() error {
	if len(dbab.BoardIDs) == 0 {
		return NoBoardsInBoardsAndBlocksErr
	}

	if len(dbab.BoardIDs) != len(dbab.BoardPatches) {
		return BoardIDsAndPatchesMissmatchInBoardsAndBlocksErr
	}

	if len(dbab.BlockIDs) == 0 {
		return NoBlocksInBoardsAndBlocksErr
	}

	if len(dbab.BlockIDs) != len(dbab.BlockPatches) {
		return BlockIDsAndPatchesMissmatchInBoardsAndBlocksErr
	}

	return nil
}

func GenerateBoardsAndBlocksIDs(bab *BoardsAndBlocks, logger *mlog.Logger) (*BoardsAndBlocks, error) {
	if err := bab.IsValid(); err != nil {
		return nil, err
	}

	blocksByBoard := map[string][]Block{}
	for _, block := range bab.Blocks {
		blocksByBoard[block.BoardID] = append(blocksByBoard[block.BoardID], block)
	}

	boards := []*Board{}
	blocks := []Block{}
	for _, board := range bab.Boards {
		newID := utils.NewID(utils.IDTypeBoard)
		for _, block := range blocksByBoard[board.ID] {
			block.BoardID = newID
			blocks = append(blocks, block)
		}

		board.ID = newID
		boards = append(boards, board)
	}

	newBab := &BoardsAndBlocks{
		Boards: boards,
		Blocks: GenerateBlockIDs(blocks, logger),
	}

	return newBab, nil
}

func BoardsAndBlocksFromJSON(data io.Reader) *BoardsAndBlocks {
	var bab *BoardsAndBlocks
	_ = json.NewDecoder(data).Decode(&bab)
	return bab
}
