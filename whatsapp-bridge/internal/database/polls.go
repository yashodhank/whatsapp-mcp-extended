package database

// StorePollOptions persists the option list of a poll creation message, keyed by the poll
// message's own ID and chat JID. This is required to resolve incoming poll votes: WhatsApp
// only ever sends SHA-256 hashes of the selected option text, never the option text itself,
// so decrypting a vote is only half the job — the original option names have to be known
// locally to turn a hash back into something readable.
func (store *MessageStore) StorePollOptions(messageID, chatJID string, options []string) error {
	tx, err := store.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM poll_options WHERE message_id = ? AND chat_jid = ?`, messageID, chatJID); err != nil {
		return err
	}
	for i, opt := range options {
		if _, err := tx.Exec(
			`INSERT INTO poll_options (message_id, chat_jid, option_index, option_name) VALUES (?, ?, ?, ?)`,
			messageID, chatJID, i, opt,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetPollOptions returns the option names of a previously stored poll, in their original
// order. Returns an empty slice (not an error) if the poll wasn't created/seen by this bridge
// — e.g. a poll from before this feature existed, or a poll from a chat we never captured.
func (store *MessageStore) GetPollOptions(messageID, chatJID string) ([]string, error) {
	rows, err := store.db.Query(
		`SELECT option_name FROM poll_options WHERE message_id = ? AND chat_jid = ? ORDER BY option_index ASC`,
		messageID, chatJID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []string
	for rows.Next() {
		var opt string
		if err := rows.Scan(&opt); err != nil {
			return nil, err
		}
		options = append(options, opt)
	}
	return options, rows.Err()
}
