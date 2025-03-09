CREATE TABLE IF NOT EXISTS "point_history" (
	"id" INTEGER NOT NULL UNIQUE,
	"chat_id" INTEGER NOT NULL,
	"user_id" INTEGER NOT NULL,
	"amount" REAL NOT NULL,
	"source" VARCHAR(20) NOT NULL DEFAULT 'chatting',
	"change" VARCHAR(20) NOT NULL DEFAULT 'gain', -- Positive (gain) or Negative (loss)
	"timestamp" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY("id")
);

