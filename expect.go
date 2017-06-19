package main

import "log"

type Expect struct {
}

func NewExpect() *Expect {
	e := &Expect{}
	return e
}

func (e *Expect) Check(res *requestResult, ex *ExpectConfig) error {
	if res == nil {
		return ErrRequestResultIsNil
	}

	if ex == nil {
		return nil
	}

	status := ex.Status
	if status == 0 {
		status = 200
	}

	if status != res.Status {
		log.Printf("Wrong status : %d", res.Status)
		log.Printf("Expect status: %d", status)
		log.Printf("Response: %s", string(res.RawBody))
		return ErrWrongStatus
	}

	// log.Printf("request body: %s", string(res.RawBody))
	// log.Printf("expect  body: %s", ex.Body)

	if ex.Body != "" {
		err := ex.Comparator.CmpBody(res.RawBody, []byte(ex.Body))
		if err != nil {
			LogPrintfH2("Error on compare request body")
			log.Printf("Raw request body: %s", string(res.RawBody))
			log.Printf("Raw expect  body: %s", ex.Body)
			return err
		}
	}
	return nil
}
