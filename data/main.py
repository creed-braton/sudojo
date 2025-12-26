import os
import psycopg2
import json
import gzip

DB_HOST = os.environ.get("DB_HOST", "localhost")
DB_PORT = os.environ.get("DB_PORT", "5432")
DB_NAME = os.environ.get("DB_NAME", "postgres")
DB_USER = os.environ.get("DB_USER", "postgres")
DB_PASSWORD = os.environ.get("DB_PASSWORD", "password")


def pg_array(board):
    rows = ["{" + ",".join(str(cell) for cell in row) + "}" for row in board]
    return "{" + ",".join(rows) + "}"


with gzip.open("data.json.gz", "rt", encoding="utf-8") as f:
    data = json.load(f)

records = [
    (
        d["hash"],
        pg_array(d["initial_board"]),
        pg_array(d["solution"]),
        d["difficulty"],
    )
    for d in data
]

conn = psycopg2.connect(
    host=DB_HOST,
    port=DB_PORT,
    dbname=DB_NAME,
    user=DB_USER,
    password=DB_PASSWORD,
)
cur = conn.cursor()

cur.executemany(
    """
    INSERT INTO games (hash, initial_board, solution, difficulty)
    VALUES (%s, %s::int[], %s::int[], %s::difficulty)
    ON CONFLICT (hash) DO NOTHING
""",
    records,
)

conn.commit()
cur.close()
conn.close()
