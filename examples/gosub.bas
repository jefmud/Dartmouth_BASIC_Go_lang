10 REM Subroutine Example
20 PRINT "Main program start"
30 GOSUB 100
40 PRINT "Back in main"
50 GOSUB 100
60 PRINT "Done"
70 END
100 REM Subroutine
110 PRINT "Inside subroutine"
120 RETURN
