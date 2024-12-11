package tempdatahandlers

// var temporaryData = make(map[int64]map[string]string)

// Установка временных данных
func SetTemporaryData(userID int64, key, value string, temporaryData map[int64]map[string]string) {
	if _, exists := temporaryData[userID]; !exists {
		temporaryData[userID] = make(map[string]string)
	}
	temporaryData[userID][key] = value
}

// Получение временных данных
func GetTemporaryData(userID int64, temporaryData map[int64]map[string]string) map[string]string {
	if data, exists := temporaryData[userID]; exists {
		return data
	}
	return make(map[string]string)
}

// Удаление временных данных
func DeleteTemporaryData(userID int64, temporaryData map[int64]map[string]string) {
	delete(temporaryData, userID)
}
