package disk

// filterDrives returns all drives for every action — drives are shown in
// every list. Selection-time guards in updateDriveList block operations on
// ineligible drives (root, chunk, already-mounted, etc.) with a clear message.
func filterDrives(drives []Drive, action actionType) []Drive {
	return drives
}
