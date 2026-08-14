package main

import (
	"fmt"
	"sync"
)

type Account struct {
	balance      int
	pendingHolds int
}

func writer(acct *Account, uncommittedWrite chan struct{}, readerCommitted chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	writerChange := 50
	acct.pendingHolds -= writerChange

	fmt.Printf("Writer uncommitted write: %d\n", acct.pendingHolds)

	uncommittedWrite <- struct{}{}
	<-readerCommitted

	acct.pendingHolds += writerChange

	fmt.Printf("Writer rolls back:%d\n", acct.pendingHolds)
}

func reader(acct *Account, uncommittedWrite chan struct{}, readerCommitted chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	<-uncommittedWrite

	if acct.pendingHolds != 15 {

		fmt.Printf("Writer's release never landed:%d\n", acct.pendingHolds)
	}

	readerCommit := 80
	acct.pendingHolds += readerCommit
	fmt.Printf("Reader committed:%d\n", acct.pendingHolds)

	readerCommitted <- struct{}{}

}
func main() {

	acct := &Account{balance: 100, pendingHolds: 65}

	chan1 := make(chan struct{})
	chan2 := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(2)

	go writer(acct, chan1, chan2, &wg)
	go reader(acct, chan1, chan2, &wg)

	wg.Wait()

	fmt.Printf("Final Balance:%d\n", acct.balance-acct.pendingHolds)

}
