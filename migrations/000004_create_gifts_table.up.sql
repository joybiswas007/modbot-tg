CREATE TABLE IF NOT EXISTS gifts (
    id INTEGER NOT NULL UNIQUE,
    chat_id INTEGER NOT NULL,
    sender_id INTEGER NOT NULL,   -- The user who is sending the gift
    receiver_id INTEGER NOT NULL, -- The user who is receiving the gift
    amount INTEGER NOT NULL,      -- Points transferred
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
    FOREIGN KEY(sender_id) REFERENCES users(user_id),
    FOREIGN KEY(receiver_id) REFERENCES users(user_id),
    PRIMARY KEY("ID")
);
