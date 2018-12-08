create table product_category (
  product_id text not null,
  category_id text not null,
  primary key(product_id, category_id)
);
