package mysqldb

func (model *Model) Begin() error {
	model.showTransaction("BEGIN")
	if model.isAutoCommit {
		tx, err := model.db.Begin()
		if err != nil {
			return err
		}
		model.isAutoCommit = false
		model.isExecuted = false
		model.tx = tx
	}
	return nil
}

func (model *Model) Rollback() error {
	model.showTransaction("ROLLBACK")
	if !model.isAutoCommit && !model.isExecuted {
		model.isAutoCommit = true
		model.isExecuted = true
		return model.tx.Rollback()
	}
	return nil
}

func (model *Model) Commit() error {
	model.showTransaction("COMMIT")
	if !model.isAutoCommit && !model.isExecuted {
		model.isAutoCommit = true
		model.isExecuted = true
		return model.tx.Commit()
	}

	return nil
}

func (model *Model) Close() {
	if model.db != nil {
		if model.tx != nil && !model.isExecuted {
			model.Rollback()
		}
		model.tx = nil
		model.db = nil
	}
}

func (model *Model) showTransaction(args string) {
	if model.adapter.isLog {
		model.adapter.logger.Debugf("%-8s TX %v ", args, &model.tx)
	}
}
