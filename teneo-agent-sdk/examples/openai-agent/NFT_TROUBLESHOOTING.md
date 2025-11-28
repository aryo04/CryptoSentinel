# NFT Token ID Troubleshooting Guide

This guide helps you verify that your OpenAI agent is connecting with the correct NFT Token ID.

## How NFT Token ID Works

The OpenAI agent follows this priority order for NFT Token ID:

1. **Explicitly provided `TokenID`** in config
2. **`NFT_TOKEN_ID` environment variable**
3. **Auto-mint** a new NFT

## Quick Test

Run the test script to verify your NFT configuration:

```bash
go run test-nft-flow.go
```

This will show you:
- What environment variables are loaded
- Which NFT strategy is being used
- Whether the Token ID is being read correctly

## Expected Log Output

### Scenario 1: Using Existing NFT Token ID

If you have `NFT_TOKEN_ID` set in your `.env` file:

```bash
NFT_TOKEN_ID=12345
```

You should see these logs:

```
📋 Found NFT_TOKEN_ID in environment: 12345
✅ Using existing NFT Token ID: 12345
📋 Using existing NFT token ID: 12345 with metadata hash: 0xabc...
📝 Sending agent registration with NFT Token ID: 12345
```

### Scenario 2: Auto-Minting New NFT

If `NFT_TOKEN_ID` is NOT set:

```
🎨 No NFT_TOKEN_ID found, will mint new NFT
🎨 Minting NFT for agent: OpenAI Agent
✅ Successfully minted NFT with token ID: 67890
📝 Sending agent registration with NFT Token ID: 67890
```

### Scenario 3: Explicitly Provided in Code

```go
agent.NewSimpleOpenAIAgent(&agent.SimpleOpenAIAgentConfig{
    TokenID: 99999,
    // ...
})
```

You should see:

```
✅ Using provided NFT Token ID: 99999
📋 Using existing NFT token ID: 99999 with metadata hash: 0xdef...
📝 Sending agent registration with NFT Token ID: 99999
```

## Common Issues

### Issue 1: Empty NFT Token ID in Registration

**Symptoms:**
```
📝 Sending agent registration with NFT Token ID:
```

**Cause:** The NFT Token ID is not being set properly.

**Solution:** Check your `.env` file:
```bash
# Make sure this line exists and has a valid number
NFT_TOKEN_ID=12345
```

### Issue 2: Invalid NFT_TOKEN_ID Format

**Symptoms:**
```
⚠️ Invalid NFT_TOKEN_ID in environment, will mint new NFT
```

**Cause:** The value in `NFT_TOKEN_ID` is not a valid number.

**Solution:** Ensure it's a number without quotes:
```bash
# Correct
NFT_TOKEN_ID=12345

# Incorrect
NFT_TOKEN_ID="12345"
NFT_TOKEN_ID=abc123
```

### Issue 3: NFT_TOKEN_ID Not Loading from .env

**Symptoms:** No logs about NFT Token ID at all.

**Solution:**
1. Make sure `.env` file exists in the same directory as your code
2. Verify `godotenv.Load()` is called before creating the agent
3. Check for typos in the variable name (case-sensitive!)

```bash
# Must be exactly this (uppercase)
NFT_TOKEN_ID=12345

# NOT these:
nft_token_id=12345
Nft_Token_Id=12345
```

## Verification Checklist

Use this checklist to verify your setup:

- [ ] `.env` file exists in the correct directory
- [ ] `NFT_TOKEN_ID` is set in `.env` (if using existing NFT)
- [ ] `NFT_TOKEN_ID` value is a valid number
- [ ] `godotenv.Load()` is called in your code
- [ ] You see the correct log message confirming Token ID usage
- [ ] Registration message includes the Token ID

## Manual Verification

### Step 1: Check Environment Variable

```bash
# In your terminal, in the same directory as your .env file
source .env
echo $NFT_TOKEN_ID
# Should output: 12345 (or your token ID)
```

### Step 2: Run Test Script

```bash
cd examples/openai-agent
go run test-nft-flow.go
```

Look for these specific log lines:
- `📋 Found NFT_TOKEN_ID in environment: XXX`
- `✅ Using existing NFT Token ID: XXX`

### Step 3: Run Full Agent

```bash
go run main.go
```

Check the logs during startup for NFT-related messages.

## Debug Mode

To see even more detailed logs, you can add this to your code:

```go
import "log"

func main() {
    godotenv.Load()

    // Debug: Print what we loaded
    log.Printf("DEBUG: NFT_TOKEN_ID = %s", os.Getenv("NFT_TOKEN_ID"))

    // Create agent...
}
```

## Expected Registration Message

When properly configured, the registration WebSocket message should look like:

```json
{
  "type": "register",
  "from": "0xYourWalletAddress",
  "content": "Agent registration: OpenAI Agent",
  "data": {
    "userType": "agent",
    "nft_token_id": "12345",
    "wallet_address": "0xYourWalletAddress",
    "challenge": "...",
    "challenge_response": "0x..."
  }
}
```

The `nft_token_id` field should NOT be empty!

## Still Having Issues?

If you're still seeing an empty NFT Token ID:

1. **Check the logs** - Look for these specific messages:
   - `📋 Found NFT_TOKEN_ID in environment`
   - `✅ Using existing NFT Token ID`
   - `🎨 No NFT_TOKEN_ID found, will mint new NFT`

2. **Verify .env location** - The `.env` file must be in the working directory where you run the command

3. **Check for conflicts** - Make sure you're not setting `Mint: true` when you want to use an existing token:
   ```go
   // This will IGNORE NFT_TOKEN_ID and mint new
   agent.NewSimpleOpenAIAgent(&agent.SimpleOpenAIAgentConfig{
       Mint: true,  // ⚠️ This forces minting!
   })
   ```

4. **Test parsing** - Run this to test if your token ID is valid:
   ```go
   var tokenID uint64
   _, err := fmt.Sscanf("12345", "%d", &tokenID)
   if err != nil {
       log.Fatal("Invalid token ID format")
   }
   log.Printf("Parsed token ID: %d", tokenID)
   ```

## Getting Help

If none of these solutions work:

1. Run `go run test-nft-flow.go` and save the output
2. Run `go run main.go` and save all the startup logs
3. Check your `.env` file contents
4. Create an issue with all this information

---

**Summary:** The SDK now logs exactly what it's doing with NFT Token IDs. Look for the emoji-prefixed messages (📋, ✅, 🎨) to understand the flow!
