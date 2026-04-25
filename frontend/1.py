#!/usr/bin/env python3
import sys
import time

def slow_print(line, delay=0.03):
    for ch in line:
        sys.stdout.write(ch)
        sys.stdout.flush()
        time.sleep(delay)
    sys.stdout.write("\n")
    sys.stdout.flush()

def bar(label, filled, total=20):
    blocks = "█" * filled + "-" * (total - filled)
    return f"[BAR]    {label:<12} [{blocks}] {int(filled/total*100):>3}%"

slow_print("$ python3 linuxdo_2nd_anniv.py\n", 0.02)

slow_print("[BOOT]   initializing linux.do anniversary sequence...")
time.sleep(0.4)
slow_print("[OK]     uptime: 2 years 0 days")
slow_print("[OK]     nodes online: 38000+")
slow_print("[OK]     mood: 99.99% happy\n")

slow_print("[SCAN]   scanning timeline for classic posts...")
time.sleep(0.4)
slow_print("[FOUND]  memes:         999+")
slow_print("[FOUND]  serious posts: 42")
slow_print("[FOUND]  pure_water:    infinite\n")

for filled in (5, 10, 15, 18):
    slow_print(bar("loading YEAR 1", filled))
    time.sleep(0.2)
slow_print("")

for filled in (8, 12, 16, 18):
    slow_print(bar("loading YEAR 2", filled))
    time.sleep(0.2)
slow_print("")

for filled in (10, 14, 18, 20):
    slow_print(bar("loading YEAR 3", filled))
    time.sleep(0.2)
slow_print("")

slow_print("[LOG]    role bonuses:")
slow_print("         - old_boy      +10% memory")
slow_print("         - lurker       +10% courage")
slow_print("         - newbie       +10% luck")
slow_print("         - maintainer   +10% patience\n")

slow_print("[EVENT]  HAPPY 2nd ANNIVERSARY, linux.do!")
slow_print("         may uptime == infinite;")
slow_print("         may bugs   == solvable;")
slow_print("         may friends++ every day;\n")

slow_print("[END]    celebration completed. press any key to continue_")
