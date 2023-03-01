CREATE TABLE tblCRQueryQueue
(
    pkQueryQueueID INT AUTO_INCREMENT PRIMARY KEY,
    runStatus ENUM ('pending', 'processing', 'completed', 'failed') DEFAULT 'pending' NOT NULL,
    runError TEXT NULL,
    runTime INT DEFAULT 0 NULL,
    runRepeat VARCHAR(50) NULL,
    runFirst DATETIME DEFAULT CURRENT_TIMESTAMP NULL,
    runLast DATETIME NULL,
    runNext DATETIME NULL,
    queryName TINYTEXT NOT NULL,
    querySignature VARCHAR(35) NOT NULL
);