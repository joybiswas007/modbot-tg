package main

// Error messages separated for easy modification
const (
	ErrNoStatsFound       = "**No stats found for this user.**"
	ErrNoRankingsYet      = "**No rankings yet! Try chatting more to earn points.**"
	ErrNoHistoryFound     = "**No history found for user!**"
	ErrInvalidGiftAmount  = "Invalid gift amount. It must be a positive number."
	ErrInvalidID          = "Invalid ID format."
	ErrNoItemFound        = "**No such item found with the specified id!**"
	ErrGiftToSelf         = "You can't send gift to yourself!"
	ErrGiftToBot          = "No need to gift to bot"
	ErrGiftAmountZero     = "Gift amount must be greater than 0."
	ErrNoPointsToGift     = "**You don’t have enough points to gift!**"
	ErrUserDoesNotExist   = "User doesn't exist!"
	ErrGiftProcessingFail = "Failed to process the gift. Please try again later."
	ErrNoBoost            = "🚫 **No active boosts found!**"
	ErrNoItem             = "**Looks like there is no item(s) in shop yet!**"
	ErrNoPointsYet        = "**You don’t have any points yet! Start chatting to earn some...**"
	ErrNotEnoughPoints    = "**Insufficient points! Start chatting to earn some...**"
	ErrBuyingItem         = "**Failed to buy item!**"
	ErrUnknown            = "**Something went wrong!**"
)
