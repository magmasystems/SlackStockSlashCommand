
-- Clear the tables
DELETE FROM slackstockbot.alertsubscription WHERE id > 0;
DELETE FROM slackstockbot.stockprice WHERE symbol IS NOT NULL;

INSERT INTO slackstockbot.stockprice(symbol, price, time) VALUES
('MSFT', 133.01, current_date),
('INTC', 45.32, current_date),
('GE', 9.45, current_date),
('UBER', 52.19, current_date);

INSERT INTO slackstockbot.alertsubscription(slackuser, webhook, symbol, targetprice, direction) VALUES
('Marc Adler', 'https://hooks.slack.com/commands/TKDT0R4PQ/652850474995/', 'MSFT', 130.0, 'ABOVE'),
('Marc Adler', 'https://hooks.slack.com/commands/TKDT0R4PQ/652850474995/', 'INTC', 45.0, 'BELOW'),
('Marc Adler', 'https://hooks.slack.com/commands/TKDT0R4PQ/652850474995/', 'GE', 10.0, 'BELOW'),
('Marc Adler', 'https://hooks.slack.com/commands/TKDT0R4PQ/652850474995/', 'UBER', 50.0, 'ABOVE');



SELECT a.slackuser, a.webhook, a.symbol, a.targetprice, a.direction, p.price
	FROM slackstockbot.alertsubscription a, slackstockbot.stockprice p
	WHERE a.wasnotified = false AND a.symbol = p.symbol AND
		( (a.direction = 'ABOVE' AND p.price >= a.targetprice) OR (a.direction = 'BELOW' AND p.price <= a.targetprice) )

