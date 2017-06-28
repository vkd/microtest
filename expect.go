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
	if status == 0 && res.Status != 0 {
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

	var body string

	if ex.Body != "" {
		body = ex.Body
	} else if ex.BodyMin != "" {
		ex.Comparator.IsLeast = true
		body = ex.BodyMin
	}

	if len(body) > 0 {
		err := ex.Comparator.CmpBody(res.RawBody, []byte(body))
		if err != nil {
			LogPrintfH2("Error on compare request body")
			log.Printf("Raw request body: %s", string(res.RawBody))
			log.Printf("Raw expect  body: %s", body)
			return err
		}
	}

	return nil
}
