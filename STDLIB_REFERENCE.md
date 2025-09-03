# 📚 Jabline Standard Library Reference

**Quick reference for all built-in functions and modules in Jabline v2.0.0**

## 🔧 Built-in Functions

### File System
```jabline
readFile(filename)              // Read file contents as string
writeFile(filename, content)    // Write string to file
fileExists(filename)            // Check if file exists (boolean)
deleteFile(filename)            // Delete file
createDir(dirname)              // Create directory
listDir(dirname)                // List directory contents (array of objects)
getWorkingDir()                 // Get current working directory
changeDir(dirname)              // Change working directory
```

### Path Utilities
```jabline
pathJoin(...paths)              // Join path segments
pathBase(path)                  // Extract filename from path
pathDir(path)                   // Extract directory from path
```

### Network
```jabline
httpGet(url)                    // HTTP GET request (returns {status, body})
httpPost(url, data)             // HTTP POST request (returns {status, body})
```

### Environment
```jabline
getEnv(key)                     // Get environment variable (null if not found)
setEnv(key, value)              // Set environment variable
```

### Time
```jabline
now()                           // Current timestamp (integer)
formatTime(timestamp, format)   // Format timestamp (YYYY-MM-DD HH:mm:ss)
sleep(milliseconds)             // Pause execution
```

### Console
```jabline
print(...args)                  // Print without newline
println(...args)                // Print with newline
echo(value)                     // Print value with newline (built-in)
```

### Core Functions
```jabline
len(array|string|hash)          // Get length/size
type(value)                     // Get type as string
push(array, value)              // Add element to array
charAt(string, index)           // Get character at index
substring(string, start, end)   // Extract substring
indexOf(string, search)         // Find index of substring
lastIndexOf(string, search)     // Find last index of substring
upper(string)                   // Convert to uppercase
lower(string)                   // Convert to lowercase
trim(string)                    // Remove whitespace
```

## 📦 Standard Library Modules

### Math Module (`math`)
```jabline
import { abs, max, min, pow, factorial } from "math";

// Constants
PI = 3                          // Pi constant (simplified)
E = 2                           // Euler's number (simplified)

// Functions
abs(x)                          // Absolute value
max(a, b)                       // Maximum of two numbers
min(a, b)                       // Minimum of two numbers
pow(base, exponent)             // Power operation
sqrt(x)                         // Square root (approximation)
factorial(n)                    // Factorial
isEven(n)                       // Check if number is even
isOdd(n)                        // Check if number is odd
random(min, max)                // Random number (basic)
round(x)                        // Round number
floor(x)                        // Floor operation
ceil(x)                         // Ceiling operation
```

### Strings Module (`strings_minimal`)
```jabline
import { capitalize, formatName, isValidEmail } from "strings_minimal";

// Constants
MESSAGE = "Strings module loaded successfully"

// Functions
capitalize(str)                 // Capitalize first letter
formatName(first, last)         // Format full name
isValidEmail(email)             // Basic email validation
extractDomain(email)            // Extract domain from email
cleanSpaces(str)                // Remove extra whitespace
```

### Arrays Module (`arrays`)
```jabline
import { findIndex, sum, sort, max, min } from "arrays";

// Search and filter
findIndex(arr, value)           // Find index of value
contains(arr, value)            // Check if array contains value
remove(arr, value)              // Remove first occurrence of value
removeAt(arr, index)            // Remove element at index

// Manipulation
insertAt(arr, index, value)     // Insert value at index
slice(arr, start, end)          // Extract portion of array
concat(arr1, arr2)              // Concatenate two arrays
reverseArray(arr)               // Reverse array order
unique(arr)                     // Remove duplicates
flatten(arr)                    // Flatten nested arrays
shuffle(arr)                    // Shuffle array elements
chunk(arr, size)                // Split into chunks

// Mathematics
sum(arr)                        // Sum all elements
average(arr)                    // Calculate average
max(arr)                        // Find maximum value
min(arr)                        // Find minimum value
sort(arr)                       // Sort array (numbers)

// Validation
all(arr, value)                 // Check if all elements equal value
any(arr, value)                 // Check if any element equals value
countValue(arr, value)          // Count occurrences of value
```

### Date/Time Module (`time/datetime`)
```jabline
import { createDate, formatDate, isLeapYear } from "time/datetime";

// Constants
SECONDS_IN_MINUTE = 60
MINUTES_IN_HOUR = 60
HOURS_IN_DAY = 24
MONTH_NAMES = ["Enero", "Febrero", ...]
DAY_NAMES = ["Domingo", "Lunes", ...]

// Date creation and validation
createDate(day, month, year)    // Create date object
isLeapYear(year)                // Check if year is leap year
addDays(date, days)             // Add days to date (simplified)

// Formatting
formatDate(date)                // Format as DD/MM/YYYY
getMonthName(date)              // Get month name

// Time operations
createTime(hours, mins, secs)   // Create time object
formatTime(time)                // Format as HH:MM:SS
createDateTime(d,m,y,h,mi,s)    // Create combined date-time
formatDateTime(datetime)        // Format date and time
```

### JSON Module (`data/json`)
```jabline
import { stringify, parse, isValid } from "data/json";

// Core functions
stringify(obj)                  // Convert object to JSON string
parse(jsonStr)                  // Parse JSON string to object
isValid(jsonStr)                // Check if JSON is valid
prettify(obj)                   // Format JSON with indentation (basic)
```

### Testing Module (`testing/assert`)
```jabline
import { describe, it, assertEqual, assertTrue, showFinalReport } from "testing/assert";

// Test structure
describe(name, testFunction)    // Define test suite
it(name, testFunction)         // Define individual test

// Assertions
assertEqual(actual, expected, msg)      // Check equality
assertNotEqual(actual, expected, msg)   // Check inequality
assertTrue(condition, msg)              // Check if true
assertFalse(condition, msg)             // Check if false
assertNull(value, msg)                  // Check if null
assertNotNull(value, msg)               // Check if not null
assertType(value, expectedType, msg)    // Check type
assertLength(value, expectedLen, msg)   // Check length
assertGreaterThan(actual, expected, msg) // Check greater than
assertLessThan(actual, expected, msg)   // Check less than

// Test management
resetStats()                    // Reset test statistics
getStats()                      // Get current test stats
showFinalReport()               // Display final test report
```

### Collections Module (`data/collections`)
```jabline
import { map, filter, reduce, groupBy } from "data/collections";

// Functional operations
map(arr, mapFn)                 // Transform each element
filter(arr, filterFn)           // Filter elements
reduce(arr, reduceFn, initial)  // Reduce to single value
find(arr, findFn)               // Find first matching element
findIndex(arr, findFn)          // Find index of first match
every(arr, testFn)              // Check if all elements pass test
some(arr, testFn)               // Check if any element passes test

// Array utilities
chunk(arr, size)                // Split into chunks
flatten(arr)                    // Flatten nested arrays
zip(arr1, arr2)                 // Combine arrays element-wise
unzip(pairsArray)               // Separate paired elements

// Set operations
union(arr1, arr2)               // Union of two arrays
intersection(arr1, arr2)        // Common elements
difference(arr1, arr2)          // Elements in arr1 not in arr2

// Object operations
mapObject(obj, mapFn)           // Transform object values
filterObject(obj, filterFn)     // Filter object properties
pickKeys(obj, keys)             // Select specific keys
omitKeys(obj, keys)             // Exclude specific keys

// Grouping
groupBy(arr, keyFn)             // Group elements by key function
countBy(arr, keyFn)             // Count elements by key function
sortBy(arr, sortFn)             // Sort by function result
partition(arr, testFn)          // Split into two arrays

// Pipeline
pipe(value, functions)          // Execute functions in sequence
compose(functions)              // Compose functions (reverse order)
```

### Crypto Module (`crypto/hash`)
```jabline
import { simpleHash, base64Encode, generateToken } from "crypto/hash";

// Constants
SHA256_DIGEST_SIZE = 32
MD5_DIGEST_SIZE = 16
BASE64_CHARS = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

// Hashing
simpleHash(input)               // Simple hash function (32-bit)
md5(input)                      // MD5 hash (simulated)
sha256(input)                   // SHA256 hash (simulated)

// Encoding
base64Encode(input)             // Encode to Base64
base64Decode(input)             // Decode from Base64
isValidBase64(input)            // Check if valid Base64

// Security utilities
safeCompare(str1, str2)         // Timing-safe string comparison
generateToken(length)           // Generate random token
generateUUID()                  // Generate UUID v4 (simulated)
checksum(input)                 // Simple checksum
verifyChecksum(input, expected) // Verify checksum
```

### OS/Environment Module (`os/env`)
```jabline
import { getUserInfo, detectOS, createConfig } from "os/env";

// Environment
getEnvWithDefault(key, default)  // Get env var with fallback
hasEnv(key)                      // Check if env var exists
getCommonEnvVars()               // Get common system vars
setMultipleEnv(envMap)           // Set multiple env vars

// System detection
detectOS()                       // Detect OS ("windows", "unix", "unknown")
isWindows()                      // Check if Windows
isUnix()                         // Check if Unix/Linux

// User information
getUserInfo()                    // Get current user info
getSystemPaths()                 // Get important system paths

// Configuration
createConfig()                   // Create app config from env
loadConfigWithPrefix(prefix)     // Load config with env prefix
validateRequiredEnv(vars)        // Validate required env vars
validateAppConfig()              // Validate app configuration

// Debugging
debugEnvVars()                   // Show common env vars
getSystemReport()                // Complete system report
```

## 💡 Usage Examples

### Complete Application Example
```jabline
import { stringify, parse } from "data/json";
import { sum, max } from "arrays";
import { capitalize } from "strings_minimal";

// Load configuration
fn loadConfig() {
    if (!fileExists("config.json")) {
        let defaultConfig = {
            "name": "My App",
            "port": getEnv("PORT") ?? "3000",
            "debug": false
        };
        writeFile("config.json", stringify(defaultConfig));
        return defaultConfig;
    }
    
    let content = readFile("config.json");
    return parse(content);
}

// Process data
fn processUsers(users) {
    let processed = [];
    for (user in users) {
        let processedUser = {
            "name": capitalize(user["name"]),
            "email": user["email"]
        };
        processed = push(processed, processedUser);
    }
    return processed;
}

// Main application
fn main() {
    let config = loadConfig();
    echo("Starting " + config["name"]);
    
    // Simulate processing
    let users = [
        {"name": "john", "email": "john@example.com"},
        {"name": "jane", "email": "jane@example.com"}
    ];
    
    let processed = processUsers(users);
    writeFile("output.json", stringify(processed));
    
    echo("Processed " + len(users) + " users");
}

main();
```

---

**Jabline Standard Library v2.0.0** - Complete reference for production development