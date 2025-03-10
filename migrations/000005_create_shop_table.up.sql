CREATE TABLE IF NOT EXISTS shop (
    id INTEGER NOT NULL UNIQUE, 
    name TEXT NOT NULL UNIQUE,
    type VARCHAR(20) NOT NULL,
    description TEXT NOT NULL,
    price REAL NOT NULL,
    duration INTEGER NOT NULL DEFAULT 0, -- Duration in hours (0 if not time-based)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY("id")
);


INSERT INTO shop (name, type, description, price, duration) 
VALUES 
    ('Double Points', 'double_points', 'Earn double points for 12 hours.', 10000, 12),
    ('Lucky Bonus', 'lucky_bonus', 'Chance to get 10-50% extra points for 6 hours.', 7500, 6);


