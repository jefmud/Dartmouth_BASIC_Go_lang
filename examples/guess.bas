10 REM Guess the Number Game
20 LET N = 42
30 PRINT "I'm thinking of a number between 1 and 100"
40 INPUT "Enter your guess: "; G
50 IF G == N THEN GOTO 100
60 IF G < N THEN PRINT "Too low!"
70 IF G > N THEN PRINT "Too high!"
80 GOTO 40
100 PRINT "Correct! You guessed it!"
110 END
