package daityou

type DaityouCell struct {
	AuthorUser   string //貸主
	RentaledUser string //借主
	Amount       uint   //金額
}

func (cell *DaityouCell) AddBalance(amount uint) {
	cell.Amount += amount
}

type DaityouAuthorsMap map[string]*DaityouCell

// ------DaityouManager------
type DaityouManager struct {
	DaityouMap map[string]DaityouAuthorsMap
}

func NewDaityouManager() *DaityouManager {
	return &DaityouManager{
		DaityouMap: make(map[string]DaityouAuthorsMap),
	}
}

func (manager *DaityouManager) GetDaityouCell(authorName string, rentaledUserName string) *DaityouCell {
	_, ok := manager.DaityouMap[authorName]
	if !ok {
		manager.DaityouMap[authorName] = make(DaityouAuthorsMap)
	}

	authorMap := manager.DaityouMap[authorName]

	if _, ok := authorMap[rentaledUserName]; !ok {
		authorMap[rentaledUserName] = &DaityouCell{
			AuthorUser:   authorName,
			RentaledUser: rentaledUserName,
			Amount:       0,
		}
	}

	cell := authorMap[rentaledUserName]
	return cell
}

func (manager *DaityouManager) EasyPay(authorName string, rentaledUserName string, amount uint) {
	cell := manager.GetDaityouCell(authorName, rentaledUserName)
	cell.AddBalance(amount)
}

func (manager *DaityouManager) CalcState(userName1 string, userName2 string) DaityouCell {
	cell1 := manager.GetDaityouCell(userName1, userName2)
	cell2 := manager.GetDaityouCell(userName2, userName1)

	if cell1.Amount >= cell2.Amount {
		return DaityouCell{
			AuthorUser:   userName1,
			RentaledUser: userName2,
			Amount:       cell1.Amount - cell2.Amount,
		}
	} else {
		return DaityouCell{
			AuthorUser:   userName2,
			RentaledUser: userName1,
			Amount:       cell2.Amount - cell1.Amount,
		}
	}
}

// userが貸主となる全てのセルを取得
func (manager *DaityouManager) GetAuthorStates(userName string) []DaityouCell {
	var results []DaityouCell
	authorMap := manager.DaityouMap[userName]

	for _, cell := range authorMap {
		state := manager.CalcState(userName, cell.RentaledUser)

		if state.AuthorUser == userName {
			results = append(results, state)
		}
	}

	return results
}
