10 REM Simple text lunar lander
20 LET ALT = 1000
30 LET VEL = -50
40 LET FUEL = 200
50 LET GRAV = -0.16
55 GOSUB 500
60 PRINT "LUNAR LANDER - BEGIN DESCENT"
70 PRINT "ALT=", ALT, " VEL=", VEL, " FUEL=", FUEL
80 INPUT "THRUST (0-50)? "; T
90 IF T < 0 THEN LET T = 0
100 IF T > 50 THEN LET T = 50
110 IF FUEL <= 0 THEN LET T = 0
120 LET BURN = T
130 IF BURN > FUEL THEN LET BURN = FUEL
140 LET FUEL = FUEL - BURN
150 LET ACC = GRAV + (BURN / 2)
160 LET VEL = VEL + ACC
170 LET ALT = ALT + VEL
180 IF ALT <= 0 THEN GOTO 300
190 PRINT "ALT=", ALT, " VEL=", VEL, " FUEL=", FUEL
200 GOTO 80
300 PRINT "TOUCHDOWN!"
310 IF VEL > -5 THEN PRINT "SOFT LANDING, NICE JOB": GOTO 400
320 IF VEL > -15 THEN PRINT "ROUGH LANDING, MODULE DAMAGED": GOTO 400
330 PRINT "CRASH! IMPACT VELOCITY=", VEL
400 END
500 REM Banner subroutine
505 print "*****************************************"
510 print "*****************************************"
530 print "***                                   ***"
540 print "***     L U N A R   L A N D E R       ***"
550 print "***                                   ***"
560 print "***              __                   ***"
580 print "***             /  \                  ***"
590 print "***            /____\                 ***"
600 print "***            |    |                 ***"
610 print "***            |NASA|                 ***"
620 print "***          __|____|__               ***"
630 print "***        _/    ][    \_             ***"
640 print "***                                   ***"
650 print "*****************************************"
660 print "*****************************************"
670 print
680 print "You are captain of a lunar landing module."
690 print "The only control is the burn rate."
700 print
702 print "Land with velocity less than 15 m/s to survive"
704 print "but extra bonus if you can get it below 5 m/s!"
706 print
710 print "  *** Can you do it Captain? ***"
720 print
730 return
