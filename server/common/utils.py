import csv
import datetime
import time


""" Bets storage location. """
STORAGE_FILEPATH = "./bets.csv"
""" Simulated winner number in the lottery contest. """
LOTTERY_WINNER_NUMBER = 7574


""" A lottery bet registry. """
class Bet:
    def __init__(self, agency: str, first_name: str, last_name: str, document: str, birthdate: str, number: str):
        """
        agency must be passed with integer format.
        birthdate must be passed with format: 'YYYY-MM-DD'.
        number must be passed with integer format.
        """
        self.agency = int(agency)
        self.first_name = first_name
        self.last_name = last_name
        self.document = document
        self.birthdate = datetime.date.fromisoformat(birthdate)
        self.number = int(number)
    
    @classmethod
    def from_csv(cls, csv_string: str):
        """Creates a Bet object from a CSV string."""
        reader = csv.reader([csv_string])
        for row in reader:
            if len(row) < 6:
                raise ValueError("Invalid CSV format for Bet")
            return cls(
                agency=row[0],
                first_name=row[1],
                last_name=row[2],
                document=row[3],
                birthdate=row[4],
                number=row[5]
            )

"""
Parse a batch of bets from the provided string.
Format: "agency,count\n" followed by count lines of bet data.
Returns a list of Bet objects.
"""
def parse_batch_bets(batch_string: str) -> list[Bet]:
    lines = batch_string.strip().split('\n')
    if len(lines) < 2:
        raise ValueError("Invalid batch format: not enough lines")
    
    # Parse header line to get agency and count
    header = lines[0].split(',')
    if len(header) != 2:
        raise ValueError("Invalid batch header format")
    
    agency = header[0]
    expected_count = int(header[1])
    
    # Check if we have the expected number of bet lines
    actual_count = len(lines) - 1  # Subtract 1 for the header line
    if actual_count != expected_count:
        raise ValueError(f"Batch format error: expected {expected_count} bets but found {actual_count}")
    
    bets = []
    for i in range(1, len(lines)):
        bet_line = lines[i]
        if not bet_line.strip():  # Skip empty lines
            continue
            
        bet_data = bet_line.split(',')
        if len(bet_data) != 5:  # FirstName,LastName,Document,Birthdate,Number
            raise ValueError(f"Invalid bet format in line {i}: {bet_line}")
        
        try:
            bet = Bet(
                agency=agency,
                first_name=bet_data[0],
                last_name=bet_data[1],
                document=bet_data[2],
                birthdate=bet_data[3],
                number=bet_data[4]
            )
            bets.append(bet)
        except Exception as e:
            raise ValueError(f"Error creating bet from line {i}: {bet_line}. Error: {str(e)}")
    
    if len(bets) != expected_count:
        raise ValueError(f"Batch processing error: expected {expected_count} valid bets but got {len(bets)}")
        
    return bets

""" Checks whether a bet won the prize or not. """
def has_won(bet: Bet) -> bool:
    return bet.number == LOTTERY_WINNER_NUMBER

"""
Persist the information of each bet in the STORAGE_FILEPATH file.
Not thread-safe/process-safe.
"""
def store_bets(bets: list[Bet]) -> None:
    with open(STORAGE_FILEPATH, 'a+') as file:
        writer = csv.writer(file, quoting=csv.QUOTE_MINIMAL)
        for bet in bets:
            writer.writerow([bet.agency, bet.first_name, bet.last_name,
                             bet.document, bet.birthdate, bet.number])

"""
Loads the information all the bets in the STORAGE_FILEPATH file.
Not thread-safe/process-safe.
"""
def load_bets() -> list[Bet]:
    with open(STORAGE_FILEPATH, 'r') as file:
        reader = csv.reader(file, quoting=csv.QUOTE_MINIMAL)
        for row in reader:
            yield Bet(row[0], row[1], row[2], row[3], row[4], row[5])

"""
Receive data from a socket until a newline character is found.
Returns the received data as a UTF-8 string without the newline character.
"""
def recv_until_newline(sock):
    buffer = b""
    while True:
        chunk = sock.recv(1024)
        if not chunk:
            # connection closed before newline
            break
        buffer += chunk
        if b"\n" in buffer:
            break
    return buffer.decode("utf-8").rstrip("\n")
    
"""
Receive data form a socket until a \0 character is found.
Returns the received data as a UTF-8 string without the \0 character.
"""
def recv_until_null(sock):
    buffer = b""
    while True:
        chunk = sock.recv(1024)
        if not chunk:
            # connection closed before null terminator
            break
        buffer += chunk
        if b"\0" in buffer:
            break
    return buffer.decode("utf-8").rstrip("\0")
