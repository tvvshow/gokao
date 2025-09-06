# CMake generated Testfile for 
# Source directory: /mnt/d/mybitcoin/gaokao/cpp-modules/device-fingerprint
# Build directory: /mnt/d/mybitcoin/gaokao/cpp-modules/device-fingerprint
# 
# This file includes the relevant testing commands required for 
# testing this directory and lists subdirectories to be tested as well.
add_test(DeviceFingerprintUnitTests "/mnt/d/mybitcoin/gaokao/cpp-modules/device-fingerprint/device_fingerprint_tests")
set_tests_properties(DeviceFingerprintUnitTests PROPERTIES  ENVIRONMENT "GTEST_OUTPUT=xml:/mnt/d/mybitcoin/gaokao/cpp-modules/device-fingerprint/test_results.xml" TIMEOUT "300" _BACKTRACE_TRIPLES "/mnt/d/mybitcoin/gaokao/cpp-modules/device-fingerprint/CMakeLists.txt;318;add_test;/mnt/d/mybitcoin/gaokao/cpp-modules/device-fingerprint/CMakeLists.txt;0;")
subdirs("_deps/googletest-build")
