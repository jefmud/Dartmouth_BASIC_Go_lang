10 REM loop test
20 print "A very basic FOR/NEXT loop"
30 FOR x = 1 to 10
40 PRINT x
50 NEXT x
60 print "Complete"
70 print "Test the step function"
80 FOR x = 1 to 10 step 2
90 print x
100 next x
110 print "Complete"
120 print "Let's test a nested loop"
130 for x = 1 to 10
140 for y = 1 to 10
150 print "x=";x;"   y=";y;"."
160 next y
170 next x
