CREATE TABLE IF NOT EXISTS storages (
	storage_id serial PRIMARY KEY,
	storage_name VARCHAR (15) UNIQUE NOT NULL,
	storage_aviable BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS products (
	product_id serial PRIMARY KEY,
	product_name VARCHAR (15),
	product_serial_code VARCHAR (20) NOT NULL,
	product_size SMALLINT CHECK (product_size >= 0),
	product_count SMALLINT CHECK (product_count >= 0)
);

CREATE TABLE IF NOT EXISTS storage_product (
	storage_id INT NOT NULL,
	product_id INT NOT NULL,
	PRIMARY KEY (storage_id, product_id),
	FOREIGN KEY (storage_id)
		REFERENCES storages (storage_id),
	FOREIGN KEY (product_id)
		REFERENCES products (product_id)
);
