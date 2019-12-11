package sqlQuery

import "database/sql"

type getGroupsStruct struct {
	id uint64
	userid sql.NullInt64
	name sql.RawBytes

}

func getGroups(tx sql.Tx, v0 int) ([]*getGroupsStruct, error) {
	res, err := tx.Query("select * from groups where id=?", v0)
	if err != nil {
		return nil, err
	}
	retgetGroupsStruct := make([]*getGroupsStruct, 0)
	for res.Next() {
		retStructRow := getGroupsStruct{}
		err = res.Scan(&retStructRow.id, &retStructRow.userid, &retStructRow.name)
		if err != nil {
			return nil, err
		}
		retgetGroupsStruct = append(retgetGroupsStruct, &retStructRow)
	}
	res.Close()
	return retgetGroupsStruct, nil 
	
}

type getUsersStruct struct {
	id uint64
	name sql.RawBytes
	info sql.RawBytes
	login sql.RawBytes
	pass sql.RawBytes
	admin uint32
	priv_key sql.RawBytes
	pub_key sql.RawBytes
	last_online uint64
	email sql.RawBytes
	enabled uint32
	code sql.RawBytes
	invited_by sql.NullInt64
	d sql.NullInt64
}

func getUsers(tx sql.Tx, v0 int) ([]*getUsersStruct, error) {
	res, err := tx.Query("select * from users where id=?", v0)
	if err != nil {
		return nil, err
	}
	retgetUsersStruct := make([]*getUsersStruct, 0)
	for res.Next() {
		retStructRow := getUsersStruct{}
		err = res.Scan(&retStructRow.id, &retStructRow.name, &retStructRow.info, &retStructRow.login, &retStructRow.pass, &retStructRow.admin, &retStructRow.priv_key, &retStructRow.pub_key, &retStructRow.last_online, &retStructRow.email, &retStructRow.enabled, &retStructRow.code, &retStructRow.invited_by, &retStructRow.d)
		if err != nil {
			return nil, err
		}
		retgetUsersStruct = append(retgetUsersStruct, &retStructRow)
	}
	res.Close()
	return retgetUsersStruct, nil 
	
}

