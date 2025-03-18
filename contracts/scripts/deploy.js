const hre = require("hardhat");

async function main() {
  // Get the contract factory
  const MultiSigWallet = await hre.ethers.getContractFactory("MultiSigWallet");

  // Get deployment parameters from environment variables
  const signerAddresses = process.env.SIGNER_ADDRESSES;
  const quorum = process.env.QUORUM;
  const recoveryAddress = process.env.RECOVERY_ADDRESS;
  const whitelistedTokens = process.env.WHITELISTED_TOKENS;

  // Validate required parameters
  if (!signerAddresses || !quorum || !recoveryAddress) {
    throw new Error("SIGNER_ADDRESSES, QUORUM, and RECOVERY_ADDRESS must be set in .env file");
  }

  // Parse signer addresses (comma-separated list)
  const signers = signerAddresses.split(',').map(addr => addr.trim());
  
  // Validate signers
  if (signers.length < 2 || signers.length > 5) {
    throw new Error("Number of signers must be between 2 and 5");
  }

  // Parse quorum
  const quorumValue = parseInt(quorum);
  
  // Validate quorum
  if (isNaN(quorumValue) || quorumValue < Math.ceil(signers.length / 2) || quorumValue > signers.length) {
    throw new Error(`Quorum must be at least ${Math.ceil(signers.length / 2)} and at most ${signers.length}`);
  }

  // Parse whitelisted tokens (comma-separated list, optional)
  const whitelistedTokensList = whitelistedTokens ? 
    whitelistedTokens.split(',').map(addr => addr.trim()) : 
    [];

  console.log("Deploying MultiSigWallet...");
  console.log("Network:", hre.network.name);
  console.log("Signers:", signers);
  console.log("Quorum:", quorumValue);
  console.log("Recovery Address:", recoveryAddress);
  console.log("Whitelisted Tokens:", whitelistedTokensList.length ? whitelistedTokensList : "None");

  // Deploy the contract
  const wallet = await MultiSigWallet.deploy(
    signers,
    quorumValue,
    recoveryAddress,
    whitelistedTokensList
  );
  await wallet.waitForDeployment();

  const address = await wallet.getAddress();
  console.log("MultiSigWallet deployed to:", address);

  // Wait for a few block confirmations
  console.log("Waiting for block confirmations...");
  await wallet.deploymentTransaction().wait(5);

  // Verify the contract
  if (hre.network.name !== "hardhat" && hre.network.name !== "localhost") {
    console.log("Verifying contract on explorer...");
    try {
      await hre.run("verify:verify", {
        address: address,
        constructorArguments: [
          signers,
          quorumValue,
          recoveryAddress,
          whitelistedTokensList
        ],
      });
      console.log("Contract verified successfully");
    } catch (error) {
      console.log("Verification failed:", error.message);
    }
  }

  return address;
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
