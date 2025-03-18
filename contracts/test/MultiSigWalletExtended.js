const { expect } = require("chai");
const { ethers } = require("hardhat");
const { time } = require("@nomicfoundation/hardhat-network-helpers");

describe("MultiSigWallet Extended Tests", function () {
  let wallet;
  let token;
  let token2; 
  let token3;
  let signers;
  let signer1;
  let signer2;
  let signer3;
  let recovery;
  let other;
  const RECOVERY_DELAY = 72 * 60 * 60; 
  const quorum = 2; // Default quorum for tests

  beforeEach(async function () {
    [signer1, signer2, signer3, recovery, other] = await ethers.getSigners();
    signers = [signer1, signer2, signer3];

    const MockToken = await ethers.getContractFactory("solidity/mocks/MockToken.sol:MockToken");
    token = await MockToken.deploy();
    token2 = await MockToken.deploy(); 
    token3 = await MockToken.deploy(); 

    const Wallet = await ethers.getContractFactory("MultiSigWallet");
    wallet = await Wallet.deploy(
      [signer1.address, signer2.address, signer3.address], // Signers array
      quorum, // Quorum (2 out of 3 needed)
      recovery.address, 
      [token.target, token2.target]
    );

    await signer1.sendTransaction({
      to: wallet.target,
      value: ethers.parseEther("10.0")
    });

    await token.transfer(wallet.target, ethers.parseUnits("100", 18));
    await token2.transfer(wallet.target, ethers.parseUnits("50", 18));
  });

  describe("Function Coverage Improvements", function () {
    it("Should verify hasReachedQuorum functions correctly", async function () {
      // Create withdrawal request
      const tx = await wallet.connect(signer1).requestWithdrawal(
        ethers.ZeroAddress,
        ethers.parseEther("1.0"),
        other.address
      );

      const receipt = await tx.wait();
      const event = receipt.logs.find(log => log.fragment && log.fragment.name === 'WithdrawalRequested');
      const requestId = event.args[0];
      
      // Check if quorum is reached (should be false with only 1 signature)
      expect(await wallet.hasReachedQuorum(requestId)).to.equal(false);
      
      // Second signer signs
      await wallet.connect(signer2).signWithdrawal(requestId);
      
      // Now quorum should be reached
      expect(await wallet.hasReachedQuorum(requestId)).to.equal(true);
    });
    
    it("Should verify if recovery address proposal has reached quorum", async function () {
      const newRecoveryAddress = other.address;
      
      // Generate the expected proposal ID
      const chainId = await ethers.provider.getNetwork().then(network => network.chainId);
      const proposalId = ethers.keccak256(
        ethers.AbiCoder.defaultAbiCoder().encode(
          ["string", "address", "uint256", "address"],
          ["RECOVERY_ADDRESS_CHANGE", newRecoveryAddress, chainId, wallet.target]
        )
      );
      
      // First signer proposes change
      await wallet.connect(signer1).proposeRecoveryAddressChange(newRecoveryAddress);
      
      // Check if quorum is reached (should be false with only 1 signature)
      expect(await wallet.hasRecoveryAddressProposalReachedQuorum(proposalId)).to.equal(false);
      
      // Second signer approves
      await wallet.connect(signer2).signRecoveryAddressChange(proposalId);
      
      // Now quorum should be reached
      expect(await wallet.hasRecoveryAddressProposalReachedQuorum(proposalId)).to.equal(true);
    });
    
    it("Should recover non-supported tokens after recovery is completed", async function () {
      // Deploy a new token that's not supported
      const MockToken = await ethers.getContractFactory("solidity/mocks/MockToken.sol:MockToken");
      const nonSupportedToken = await MockToken.deploy();
      
      // Transfer tokens to the wallet
      await nonSupportedToken.transfer(wallet.target, ethers.parseUnits("100", 18));
      
      // Request and execute recovery
      await wallet.connect(signer1).requestRecovery();
      await time.increase(RECOVERY_DELAY + 1);
      await wallet.connect(signer1).executeRecovery();
      
      // Check balances before recovery of non-supported token
      const initialBalance = await nonSupportedToken.balanceOf(wallet.target);
      expect(initialBalance).to.equal(ethers.parseUnits("100", 18));
      
      // Recover non-supported token
      await wallet.connect(signer1).recoverNonSupportedToken(nonSupportedToken.target, other.address);
      
      // Verify token was recovered
      expect(await nonSupportedToken.balanceOf(wallet.target)).to.equal(0);
      expect(await nonSupportedToken.balanceOf(other.address)).to.equal(ethers.parseUnits("100", 18));
    });
  });
  
  describe("Edge Case Coverage Improvements", function () {
    it("Should not allow adding more tokens than MAX_SUPPORTED_TOKENS", async function () {
      // First, add tokens until we approach the limit
      // By default native ETH is already in the list, so we need to add 19 more
      // Let's first check current supported tokens count
      const initialSupportedTokens = await wallet.getSupportedTokens();
      
      // Deploy and add new tokens until we reach the limit (20 tokens total including ETH)
      for (let i = 0; i < 20 - initialSupportedTokens.length; i++) {
        // Deploy new token
        const MockToken = await ethers.getContractFactory("solidity/mocks/MockToken.sol:MockToken");
        const newToken = await MockToken.deploy();
        
        // Add to supported tokens
        await wallet.connect(signer1).addSupportedToken(newToken.target);
      }
      
      // Now try to add one more token (should fail)
      const MockToken = await ethers.getContractFactory("solidity/mocks/MockToken.sol:MockToken");
      const extraToken = await MockToken.deploy();
      
      await expect(wallet.connect(signer1).addSupportedToken(extraToken.target))
        .to.be.revertedWith("Maximum supported tokens reached");
    });
    
    it("Should not allow to remove a token that is not supported", async function () {
      // Deploy new token
      const MockToken = await ethers.getContractFactory("solidity/mocks/MockToken.sol:MockToken");
      const unsupportedToken = await MockToken.deploy();
      
      await expect(wallet.connect(signer1).removeSupportedToken(unsupportedToken.target))
        .to.be.revertedWith("Token not in supported list");
    });
    
    it("Should not allow non-supported token recovery without recovery completion", async function () {
      // Deploy a new token that's not supported
      const MockToken = await ethers.getContractFactory("solidity/mocks/MockToken.sol:MockToken");
      const nonSupportedToken = await MockToken.deploy();
      
      // Transfer tokens to the wallet
      await nonSupportedToken.transfer(wallet.target, ethers.parseUnits("100", 18));
      
      // Try to recover without completing recovery
      await expect(wallet.connect(signer1).recoverNonSupportedToken(nonSupportedToken.target, other.address))
        .to.be.revertedWith("Recovery not completed");
    });
    
    it("Should not allow to recover native coin as a non-supported token", async function () {
      // Request and execute recovery
      await wallet.connect(signer1).requestRecovery();
      await time.increase(RECOVERY_DELAY + 1);
      await wallet.connect(signer1).executeRecovery();
      
      // Try to recover native coin as a non-supported token
      await expect(wallet.connect(signer1).recoverNonSupportedToken(ethers.ZeroAddress, other.address))
        .to.be.revertedWith("Cannot recover native coin");
    });
    
    it("Should not allow to recover a supported token via recoverNonSupportedToken", async function () {
      // Add token to supported tokens list
      await wallet.connect(signer1).addSupportedToken(token3.target);
      
      // Request and execute recovery
      await wallet.connect(signer1).requestRecovery();
      await time.increase(RECOVERY_DELAY + 1);
      await wallet.connect(signer1).executeRecovery();
      
      // Try to recover a supported token as a non-supported token
      await expect(wallet.connect(signer1).recoverNonSupportedToken(token3.target, other.address))
        .to.be.revertedWith("Use regular recovery for supported tokens");
    });
    
    it("Should not allow withdrawal with insufficient balance", async function () {
      // Request withdrawal with more ETH than available
      const tx = await wallet.connect(signer1).requestWithdrawal(
        ethers.ZeroAddress,
        ethers.parseEther("100.0"), // More than the 10 ETH available
        other.address
      );
      
      const receipt = await tx.wait();
      const event = receipt.logs.find(log => log.fragment && log.fragment.name === 'WithdrawalRequested');
      const requestId = event.args[0];
      
      // Second signer tries to sign (should fail on execution)
      await expect(wallet.connect(signer2).signWithdrawal(requestId))
        .to.be.revertedWith("Insufficient balance");
    });
    
    it("Should test attempt to execute a non-existent withdrawal", async function () {
      // Generate a fake request ID
      const fakeRequestId = ethers.keccak256(
        ethers.toUtf8Bytes("non-existent request")
      );
      
      // Set up a mock contract
      const MockWalletFactory = await ethers.getContractFactory("solidity/mocks/MockMultiSigWalletTest.sol:MockMultiSigWalletTest");
      const mockWallet = await MockWalletFactory.deploy(
        [signer1.address, signer2.address, signer3.address],
        quorum,
        recovery.address,
        []
      );
      
      // Try to execute non-existent request
      await expect(mockWallet.executeWithdrawalDirect(fakeRequestId))
        .to.be.revertedWith("Request not found");
    });
    
    it("Should test attempt to execute an already executed withdrawal", async function () {
      // Create withdrawal request on original wallet
      const tx = await wallet.connect(signer1).requestWithdrawal(
        ethers.ZeroAddress,
        ethers.parseEther("1.0"),
        other.address
      );
      
      const receipt = await tx.wait();
      const event = receipt.logs.find(log => log.fragment && log.fragment.name === 'WithdrawalRequested');
      const requestId = event.args[0];
      
      // Execute by getting the second signature
      await wallet.connect(signer2).signWithdrawal(requestId);
      
      // Now create a mock wallet to try to execute again
      const MockWalletFactory = await ethers.getContractFactory("solidity/mocks/MockMultiSigWalletTest.sol:MockMultiSigWalletTest");
      const mockWallet = await MockWalletFactory.deploy(
        [signer1.address, signer2.address, signer3.address],
        quorum,
        recovery.address,
        []
      );
      
      // Fund the mock wallet
      await signer1.sendTransaction({
        to: mockWallet.target,
        value: ethers.parseEther("10.0")
      });
      
      // Create a new request with the same parameters
      const mockTx = await mockWallet.connect(signer1).requestWithdrawal(
        ethers.ZeroAddress,
        ethers.parseEther("1.0"),
        other.address
      );
      
      const mockReceipt = await mockTx.wait();
      const mockEvent = mockReceipt.logs.find(log => log.fragment && log.fragment.name === 'WithdrawalRequested');
      const mockRequestId = mockEvent.args[0];
      
      // Execute it
      await mockWallet.connect(signer2).signWithdrawal(mockRequestId);
      
      // Try to execute again - should fail with "Already executed"
      // Use executeWithdrawalDirect to test the internal function directly
      await expect(mockWallet.executeWithdrawalDirect(mockRequestId))
        .to.be.revertedWith("Already executed");
    });
    
    it("Should not allow requesting recovery twice", async function () {
      // Request recovery
      await wallet.connect(signer1).requestRecovery();
      
      // Try to request again
      await expect(wallet.connect(signer2).requestRecovery())
        .to.be.revertedWith("Recovery already requested");
    });
    
    it("Should not allow requesting recovery if already executed", async function () {
      // Request and execute recovery
      await wallet.connect(signer1).requestRecovery();
      await time.increase(RECOVERY_DELAY + 1);
      await wallet.connect(signer1).executeRecovery();
      
      // Try to request recovery again - should fail
      await expect(wallet.connect(signer1).requestRecovery())
        .to.be.revertedWith("Recovery already executed");
    });
    
    it("Should revert when trying to deposit with zero amount", async function () {
      // Try to deposit 0 tokens
      await token3.approve(wallet.target, ethers.parseUnits("100", 18));
      
      await expect(wallet.depositToken(token3.target, 0))
        .to.be.revertedWith("Amount must be greater than 0");
    });
    
    it("Should not allow deposits during recovery mode", async function () {
      // Request recovery
      await wallet.connect(signer1).requestRecovery();
      
      // Try to deposit
      await token3.approve(wallet.target, ethers.parseUnits("100", 18));
      
      await expect(wallet.depositToken(token3.target, ethers.parseUnits("100", 18)))
        .to.be.revertedWith("Wallet in recovery mode");
    });
    
    it("Should not allow recovery address changes during recovery mode", async function () {
      // Request recovery
      await wallet.connect(signer1).requestRecovery();
      
      // Try to propose recovery address change
      await expect(wallet.connect(signer1).proposeRecoveryAddressChange(other.address))
        .to.be.revertedWith("Wallet in recovery mode");
    });
    
    it("Should not allow to sign recovery address change proposal twice", async function () {
      const newRecoveryAddress = other.address;
      
      // Generate the expected proposal ID
      const chainId = await ethers.provider.getNetwork().then(network => network.chainId);
      const proposalId = ethers.keccak256(
        ethers.AbiCoder.defaultAbiCoder().encode(
          ["string", "address", "uint256", "address"],
          ["RECOVERY_ADDRESS_CHANGE", newRecoveryAddress, chainId, wallet.target]
        )
      );
      
      // First signer proposes change
      await wallet.connect(signer1).proposeRecoveryAddressChange(newRecoveryAddress);
      
      // Try to sign again
      await expect(wallet.connect(signer1).signRecoveryAddressChange(proposalId))
        .to.be.revertedWith("Already signed");
    });
    
    it("Should handle complex withdrawal with insufficient token balance", async function () {
      // Create withdrawal request for more tokens than available
      const tx = await wallet.connect(signer1).requestWithdrawal(
        token.target,
        ethers.parseUnits("200", 18), // More than the 100 tokens available
        other.address
      );
      
      const receipt = await tx.wait();
      const event = receipt.logs.find(log => log.fragment && log.fragment.name === 'WithdrawalRequested');
      const requestId = event.args[0];
      
      // Second signer tries to sign (should fail on execution)
      await expect(wallet.connect(signer2).signWithdrawal(requestId))
        .to.be.revertedWith("Insufficient token balance");
    });
    
    it("Should fail when requesting withdrawal for a non-existent request", async function () {
      // Generate a fake request ID
      const fakeRequestId = ethers.keccak256(
        ethers.toUtf8Bytes("non-existent request")
      );
      
      // Try to sign non-existent request
      await expect(wallet.connect(signer1).signWithdrawal(fakeRequestId))
        .to.be.revertedWith("Request not found");
    });
    
    it("Should test for event emission on all actions", async function () {
      // Test that correct events are emitted when adding supported token
      const addTx = await wallet.connect(signer1).addSupportedToken(token3.target);
      const addReceipt = await addTx.wait();
      const addEvent = addReceipt.logs.find(log => log.fragment && log.fragment.name === 'TokenSupported');
      expect(addEvent.args[0]).to.equal(token3.target);
      
      // Test event emission when removing a token
      const removeTx = await wallet.connect(signer1).removeSupportedToken(token3.target);
      const removeReceipt = await removeTx.wait();
      const removeEvent = removeReceipt.logs.find(log => log.fragment && log.fragment.name === 'TokenRemoved');
      expect(removeEvent.args[0]).to.equal(token3.target);
      
      // Test recovery cancellation event
      await wallet.connect(signer1).requestRecovery();
      const cancelTx = await wallet.connect(signer1).cancelRecovery();
      const cancelReceipt = await cancelTx.wait();
      expect(cancelReceipt.logs.find(log => log.fragment && log.fragment.name === 'RecoveryCancelled')).to.not.be.undefined;
    });
  });
  
  describe("Constructor Validation Tests", function () {
    it("Should not allow creating wallet with fewer than 2 signers", async function () {
      const Wallet = await ethers.getContractFactory("MultiSigWallet");
      
      await expect(Wallet.deploy(
        [signer1.address], // Only 1 signer
        1,
        recovery.address,
        []
      )).to.be.revertedWith("Must have 2-7 signers");
    });
    
    it("Should not allow creating wallet with more than 7 signers", async function () {
      const [s1, s2, s3, s4, s5, s6, s7, s8] = await ethers.getSigners();
      const Wallet = await ethers.getContractFactory("MultiSigWallet");
      
      await expect(Wallet.deploy(
        [s1.address, s2.address, s3.address, s4.address, s5.address, s6.address, s7.address, s8.address], // 8 signers
        5,
        recovery.address,
        []
      )).to.be.revertedWith("Must have 2-7 signers");
    });
    
    it("Should not allow quorum less than required minimum", async function () {
      const Wallet = await ethers.getContractFactory("MultiSigWallet");
      
      await expect(Wallet.deploy(
        [signer1.address, signer2.address, signer3.address],
        1, // Quorum less than minimum (should be at least 2)
        recovery.address,
        []
      )).to.be.revertedWith("Invalid quorum");
    });
    
    it("Should not allow quorum greater than number of signers", async function () {
      const Wallet = await ethers.getContractFactory("MultiSigWallet");
      
      await expect(Wallet.deploy(
        [signer1.address, signer2.address, signer3.address],
        4, // Quorum greater than number of signers
        recovery.address,
        []
      )).to.be.revertedWith("Invalid quorum");
    });
    
    it("Should not allow zero address as recovery address", async function () {
      const Wallet = await ethers.getContractFactory("MultiSigWallet");
      
      await expect(Wallet.deploy(
        [signer1.address, signer2.address, signer3.address],
        2,
        ethers.ZeroAddress, // Zero address as recovery address
        []
      )).to.be.revertedWith("Invalid recovery address");
    });
    
    it("Should not allow duplicate signers", async function () {
      const Wallet = await ethers.getContractFactory("MultiSigWallet");
      
      await expect(Wallet.deploy(
        [signer1.address, signer1.address, signer2.address], // Duplicate signer
        2,
        recovery.address,
        []
      )).to.be.revertedWith("Duplicate signer");
    });
    
    it("Should not allow zero address as signer", async function () {
      const Wallet = await ethers.getContractFactory("MultiSigWallet");
      
      await expect(Wallet.deploy(
        [signer1.address, ethers.ZeroAddress, signer2.address], // Zero address as signer
        2,
        recovery.address,
        []
      )).to.be.revertedWith("Invalid signer address");
    });
    
    it("Should not allow whitelisting the zero address", async function () {
      const Wallet = await ethers.getContractFactory("MultiSigWallet");
      
      await expect(Wallet.deploy(
        [signer1.address, signer2.address, signer3.address],
        2,
        recovery.address,
        [ethers.ZeroAddress] // Zero address in whitelisted tokens
      )).to.be.revertedWith("Cannot whitelist zero address");
    });
  });
  
  describe("View Function Tests", function () {
    it("Should correctly report token balances", async function () {
      expect(await wallet.getTokenBalance(token.target)).to.equal(ethers.parseUnits("100", 18));
      expect(await wallet.getTokenBalance(token2.target)).to.equal(ethers.parseUnits("50", 18));
      
      // Should revert if trying to get balance of native coin
      await expect(wallet.getTokenBalance(ethers.ZeroAddress))
        .to.be.revertedWith("Use getBalance() for native coin");
    });
    
    it("Should correctly check if a signer has signed a withdrawal", async function () {
      // Create withdrawal request
      const tx = await wallet.connect(signer1).requestWithdrawal(
        ethers.ZeroAddress,
        ethers.parseEther("1.0"),
        other.address
      );
      
      const receipt = await tx.wait();
      const event = receipt.logs.find(log => log.fragment && log.fragment.name === 'WithdrawalRequested');
      const requestId = event.args[0];
      
      // Check if signers have signed
      expect(await wallet.hasSignedWithdrawal(requestId, signer1.address)).to.equal(true);
      expect(await wallet.hasSignedWithdrawal(requestId, signer2.address)).to.equal(false);
    });
    
    it("Should correctly check if a signer has signed a recovery address proposal", async function () {
      const newRecoveryAddress = other.address;
      
      // Generate the expected proposal ID
      const chainId = await ethers.provider.getNetwork().then(network => network.chainId);
      const proposalId = ethers.keccak256(
        ethers.AbiCoder.defaultAbiCoder().encode(
          ["string", "address", "uint256", "address"],
          ["RECOVERY_ADDRESS_CHANGE", newRecoveryAddress, chainId, wallet.target]
        )
      );
      
      // First signer proposes change
      await wallet.connect(signer1).proposeRecoveryAddressChange(newRecoveryAddress);
      
      // Check if signers have signed
      expect(await wallet.hasSignedRecoveryAddressProposal(proposalId, signer1.address)).to.equal(true);
      expect(await wallet.hasSignedRecoveryAddressProposal(proposalId, signer2.address)).to.equal(false);
    });
  });

  describe("Additional Uncovered Function Tests", function () {
    it("Should handle recovery address change when proposal exists", async function () {
      const newRecoveryAddress = other.address;
      
      // First signer proposes change
      await wallet.connect(signer1).proposeRecoveryAddressChange(newRecoveryAddress);
      
      // Verify recovery address has not changed (need quorum)
      expect(await wallet.recoveryAddress()).to.equal(recovery.address);
      
      // Second signer proposes the same change (adds a signature)
      await wallet.connect(signer2).proposeRecoveryAddressChange(newRecoveryAddress);
      
      // Now it should have changed after quorum is reached
      expect(await wallet.recoveryAddress()).to.equal(newRecoveryAddress);
    });
    
    it("Should handle deposit of already supported token", async function () {
      // Add token3 to supported tokens
      await wallet.connect(signer1).addSupportedToken(token3.target);
      
      // Now deposit some tokens - should work without auto-adding to supported tokens
      await token3.approve(wallet.target, ethers.parseUnits("50", 18));
      await wallet.depositToken(token3.target, ethers.parseUnits("50", 18));
      
      // Check balance
      expect(await wallet.getTokenBalance(token3.target)).to.equal(ethers.parseUnits("50", 18));
    });
    
    it("Should handle deposit of whitelisted tokens correctly", async function () {
      // token and token2 are already whitelisted
      // They should be auto-added to supported tokens when deposited
      
      // Initially not in supported tokens
      const supportedBefore = await wallet.supportedTokens(token.target);
      expect(supportedBefore).to.equal(false);
      
      // Deposit token (should auto-add to supported tokens because it's whitelisted)
      await token.approve(wallet.target, ethers.parseUnits("50", 18));
      await wallet.depositToken(token.target, ethers.parseUnits("50", 18));
      
      // Should now be in supported tokens
      const supportedAfter = await wallet.supportedTokens(token.target);
      expect(supportedAfter).to.equal(true);
    });
    
    it("Should allow non-whitelisted tokens to be deposited without auto-support", async function () {
      // Deploy a non-whitelisted token
      const MockToken = await ethers.getContractFactory("solidity/mocks/MockToken.sol:MockToken");
      const nonWhitelistedToken = await MockToken.deploy();
      
      // Deposit some tokens
      await nonWhitelistedToken.approve(wallet.target, ethers.parseUnits("50", 18));
      await wallet.depositToken(nonWhitelistedToken.target, ethers.parseUnits("50", 18));
      
      // Should not be in supported tokens
      const supported = await wallet.supportedTokens(nonWhitelistedToken.target);
      expect(supported).to.equal(false);
      
      // But balance should be updated
      expect(await wallet.getTokenBalance(nonWhitelistedToken.target)).to.equal(ethers.parseUnits("50", 18));
    });
    
    it("Should test maximum supported tokens limit for whitelisted tokens", async function () {
      // Add tokens until we reach 19 tokens (leaving 1 slot)
      for (let i = 0; i < 18; i++) {
        const MockToken = await ethers.getContractFactory("solidity/mocks/MockToken.sol:MockToken");
        const newToken = await MockToken.deploy();
        await wallet.connect(signer1).addSupportedToken(newToken.target);
      }
      
      // Create one more whitelisted token
      const MockToken = await ethers.getContractFactory("solidity/mocks/MockToken.sol:MockToken");
      const whitelistedToken = await MockToken.deploy();
      
      // Should be able to deposit and auto-add to supported tokens (using last slot)
      await whitelistedToken.approve(wallet.target, ethers.parseUnits("50", 18));
      await wallet.depositToken(whitelistedToken.target, ethers.parseUnits("50", 18));
      
      // Create another whitelisted token
      const anotherToken = await MockToken.deploy();
      
      // Try to deposit - should succeed but not auto-add to supported tokens
      await anotherToken.approve(wallet.target, ethers.parseUnits("50", 18));
      await wallet.depositToken(anotherToken.target, ethers.parseUnits("50", 18));
      
      // Should not be in supported tokens
      expect(await wallet.supportedTokens(anotherToken.target)).to.equal(false);
    });
  });
}); 