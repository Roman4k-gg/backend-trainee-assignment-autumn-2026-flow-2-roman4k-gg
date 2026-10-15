INSERT INTO users (id, name) VALUES ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'Test User') ON CONFLICT DO NOTHING;
INSERT INTO restaurants (id, name, webhook_url) VALUES ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Burger King', 'http://avito_mock_restaurant:8081/webhook') ON CONFLICT DO NOTHING;
INSERT INTO menu_items (id, restaurant_id, name, description, price, is_available) VALUES ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Whopper', 'Delicious beef burger', 300, true) ON CONFLICT DO NOTHING;
