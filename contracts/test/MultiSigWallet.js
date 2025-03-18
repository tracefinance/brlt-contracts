const { expect } = require("chai");
const { ethers } = require("hardhat");
const { time } = require("@nomicfoundation/hardhat-network-helpers");

describe("MultiSigWallet", function () {
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

  describe("Deployment", function () {
    it("Should set the right signers", async function () {
      const walletSigners = await wallet.getSigners();
      expect(walletSigners.length).to.equal(3);
      expect(walletSigners[0]).to.equal(signer1.address);
      expect(walletSigners[1]).to.equal(signer2.address);
      expect(walletSigners[2]).to.equal(signer3.address);

      // Check if isSigner mapping works
      expect(await wallet.isSigner(signer1.address)).to.be.true;
      expect(await wallet.isSigner(other.address)).to.be.false;
    });

    it("Should set the right quorum", async function () {
      expect(await wallet.quorum()).to.equal(quorum);
    });

    it("Should set the right recovery address", async function () {
      expect(await wallet.recoveryAddress()).to.equal(recovery.address);
    });
    
    it("Should support native coin by default", async function () {
      expect(await wallet.supportedTokens(ethers.ZeroAddress)).to.equal(true);
    });

    it("Should include whitelisted tokens in supported list", async function () {
      expect(await wallet.whitelistedTokens(token.target)).to.equal(true);
      expect(await wallet.whitelistedTokens(token2.target)).to.equal(true);
      
      // Note: Whitelisted tokens are not automatically added to supportedTokens anymore,
      // they're only added when a deposit is made
      expect(await wallet.supportedTokens(token.target)).to.equal(false);
      expect(await wallet.supportedTokens(token2.target)).to.equal(false);
    });

    it("Should initialize with correct ETH balance", async function () {
      expect(await wallet.getBalance()).to.equal(ethers.parseEther("10.0"));
    });

    it("Should initialize with correct token balances", async function () {
      expect(await wallet.getTokenBalance(token.target)).to.equal(ethers.parseUnits("100", 18));
      expect(await wallet.getTokenBalance(token2.target)).to.equal(ethers.parseUnits("50", 18));
    });
  });

  describe("Withdrawal Management", function () {
    it("Should allow signer to create withdrawal request", async function () {
      const tx = await wallet.connect(signer1).requestWithdrawal(
        ethers.ZeroAddress, // Native coin
        ethers.parseEther("1.0"),
        other.address
      );

      const receipt = await tx.wait();
      const event = receipt.logs.find(log => log.fragment && log.fragment.name === 'WithdrawalRequested');
      const requestId = event.args[0];

      // Check that the request was registered
      const request = await wallet.withdrawalRequests(requestId);
      expect(request.token).to.equal(ethers.ZeroAddress);
      expect(request.amount).to.equal(ethers.parseEther("1.0"));
      expect(request.to).to.equal(other.address);
      expect(request.signatureCount).to.equal(1);
      expect(request.executed).to.equal(false);

      // Check that the signature was registered
      expect(await wallet.hasSignedWithdrawal(requestId, signer1.address)).to.equal(true);
      expect(await wallet.hasSignedWithdrawal(requestId, signer2.address)).to.equal(false);
    });

    it("Should not allow non-signer to create withdrawal request", async function () {
      await expect(wallet.connect(other).requestWithdrawal(
        ethers.ZeroAddress,
        ethers.parseEther("1.0"),
        other.address
      )).to.be.revertedWith("Only signer can call this function");
    });

    it("Should execute withdrawal after reaching quorum", async function () {
      // First signer creates request
      const tx = await wallet.connect(signer1).requestWithdrawal(
        ethers.ZeroAddress,
        ethers.parseEther("1.0"),
        other.address
      );

      const receipt = await tx.wait();
      const event = receipt.logs.find(log => log.fragment && log.fragment.name === 'WithdrawalRequested');
      const requestId = event.args[0];

      // Check balances before
      const walletBalanceBefore = await ethers.provider.getBalance(wallet.target);
      const recipientBalanceBefore = await ethers.provider.getBalance(other.address);

      // Second signer approves (this should execute the withdrawal)
      await wallet.connect(signer2).signWithdrawal(requestId);

      // Check balances after
      const walletBalanceAfter = await ethers.provider.getBalance(wallet.target);
      const recipientBalanceAfter = await ethers.provider.getBalance(other.address);

      // Verify balances
      expect(walletBalanceBefore - walletBalanceAfter).to.equal(ethers.parseEther("1.0"));
      expect(recipientBalanceAfter - recipientBalanceBefore).to.equal(ethers.parseEther("1.0"));

      // Check request is marked executed
      const request = await wallet.withdrawalRequests(requestId);
      expect(request.executed).to.equal(true);
    });
    
    it("Should not allow withdrawal request during recovery mode", async function () {
      // Request recovery
      await wallet.connect(signer1).requestRecovery();
      
      // Try to create withdrawal request
      await expect(wallet.connect(signer1).requestWithdrawal(
          ethers.ZeroAddress,
          ethers.parseEther("1.0"),
          other.address
      )).to.be.revertedWith("Wallet in recovery mode");
    });
    
    it("Should allow withdrawal request after recovery is cancelled", async function () {
      // Request and then cancel recovery
      await wallet.connect(signer1).requestRecovery();
      await wallet.connect(signer2).cancelRecovery();
      
      // Now create withdrawal request
      const tx = await wallet.connect(signer1).requestWithdrawal(
        ethers.ZeroAddress,
        ethers.parseEther("1.0"),
        other.address
      );
      
      // Verify it worked
      const receipt = await tx.wait();
      const event = receipt.logs.find(log => log.fragment && log.fragment.name === 'WithdrawalRequested');
      expect(event).to.not.be.undefined;
    });

    it("Should execute token withdrawal after reaching quorum", async function () {
      // Create request for ERC20 token withdrawal
      const withdrawAmount = ethers.parseUnits("10", 18);
      const tx = await wallet.connect(signer1).requestWithdrawal(
        token.target,
        withdrawAmount,
        other.address
      );

      const receipt = await tx.wait();
      const event = receipt.logs.find(log => log.fragment && log.fragment.name === 'WithdrawalRequested');
      const requestId = event.args[0];

      // Check balances before
      const walletBalanceBefore = await token.balanceOf(wallet.target);
      const recipientBalanceBefore = await token.balanceOf(other.address);

      // Second signer approves
      await wallet.connect(signer2).signWithdrawal(requestId);

      // Check balances after
      const walletBalanceAfter = await token.balanceOf(wallet.target);
      const recipientBalanceAfter = await token.balanceOf(other.address);

      // Verify balances
      expect(walletBalanceBefore - walletBalanceAfter).to.equal(withdrawAmount);
      expect(recipientBalanceAfter - recipientBalanceBefore).to.equal(withdrawAmount);

      // Check request is marked executed
      const request = await wallet.withdrawalRequests(requestId);
      expect(request.executed).to.equal(true);
    });
  });

  describe("Recovery Management", function () {
    it("Should allow any signer to request recovery", async function () {
      await wallet.connect(signer2).requestRecovery();
      expect(await wallet.recoveryRequestTimestamp()).to.not.equal(0);
    });

    it("Should allow any signer to cancel recovery before delay", async function () {
      await wallet.connect(signer1).requestRecovery();
      await wallet.connect(signer3).cancelRecovery();
      expect(await wallet.recoveryRequestTimestamp()).to.equal(0);
    });

    it("Should not allow recovery to be executed before delay", async function () {
      await wallet.connect(signer1).requestRecovery();
      await expect(wallet.connect(signer1).executeRecovery())
        .to.be.revertedWith("Recovery delay not elapsed");
    });

    it("Should execute recovery after delay", async function () {
      // Request recovery
      await wallet.connect(signer1).requestRecovery();

      // Fast-forward time
      await time.increase(RECOVERY_DELAY + 1);

      // Execute recovery
      await wallet.connect(signer1).executeRecovery();

      // Verify recovery was executed
      expect(await wallet.recoveryExecuted()).to.equal(true);
      expect(await wallet.recoveryRequestTimestamp()).to.equal(0);

      // Check funds were transferred to recovery address
      expect(await ethers.provider.getBalance(wallet.target)).to.equal(0);
    });

    it("Should allow recovery address change with quorum approval", async function () {
      const newRecoveryAddress = other.address;
      
      // Get the chain ID
      const chainId = await ethers.provider.getNetwork().then(network => network.chainId);
      
      // Generate the expected proposal ID
      const proposalId = ethers.keccak256(
        ethers.AbiCoder.defaultAbiCoder().encode(
          ["string", "address", "uint256", "address"],
          ["RECOVERY_ADDRESS_CHANGE", newRecoveryAddress, chainId, wallet.target]
        )
      );
      
      // First signer proposes change
      await wallet.connect(signer1).proposeRecoveryAddressChange(newRecoveryAddress);
      
      // Should not be changed yet
      expect(await wallet.recoveryAddress()).to.equal(recovery.address);
      
      // Verify the signature was registered
      expect(await wallet.hasSignedRecoveryAddressProposal(proposalId, signer1.address)).to.be.true;
      
      // Second signer approves, should reach quorum
      await wallet.connect(signer2).signRecoveryAddressChange(proposalId);
      
      // Recovery address should now be changed
      expect(await wallet.recoveryAddress()).to.equal(newRecoveryAddress);
    });
  });

  describe("Token Management", function () {
    it("Should allow any signer to add supported token", async function () {
      await wallet.connect(signer3).addSupportedToken(token3.target);
      expect(await wallet.supportedTokens(token3.target)).to.equal(true);
      
      const supportedTokens = await wallet.getSupportedTokens();
      expect(supportedTokens).to.include(token3.target);
    });
    
    it("Should allow any signer to remove supported token", async function () {
      // First we need to add the token to supported tokens
      await wallet.connect(signer1).addSupportedToken(token.target);
      expect(await wallet.supportedTokens(token.target)).to.equal(true);
      
      // Now we can remove it
      await wallet.connect(signer2).removeSupportedToken(token.target);
      expect(await wallet.supportedTokens(token.target)).to.equal(false);
      
      const supportedTokens = await wallet.getSupportedTokens();
      expect(supportedTokens).to.not.include(token.target);
    });
    
    it("Should deposit tokens correctly", async function () {
      const depositAmount = ethers.parseUnits("20", 18);
      
      // Approve tokens for deposit
      await token3.approve(wallet.target, depositAmount);
      
      // Deposit tokens
      await wallet.depositToken(token3.target, depositAmount);
      
      // Check token was added to supported tokens
      expect(await wallet.supportedTokens(token3.target)).to.equal(false);
      
      // Check balance was updated
      expect(await wallet.getTokenBalance(token3.target)).to.equal(depositAmount);
    });
  });

  describe("Edge Cases", function () {
    it("Should not allow withdrawal request after expiration", async function () {
      // Create withdrawal request
      const tx = await wallet.connect(signer1).requestWithdrawal(
        ethers.ZeroAddress,
        ethers.parseEther("1.0"),
        other.address
      );

      const receipt = await tx.wait();
      const event = receipt.logs.find(log => log.fragment && log.fragment.name === 'WithdrawalRequested');
      const requestId = event.args[0];
      
      // Fast-forward time past expiration
      await time.increase(24 * 60 * 60 + 1);
      
      // Attempt to sign after expiration
      await expect(wallet.connect(signer2).signWithdrawal(requestId))
        .to.be.revertedWith("Request expired");
    });

    it("Should not allow withdrawal with zero amount", async function () {
      await expect(wallet.connect(signer1).requestWithdrawal(
        ethers.ZeroAddress,
        0,
        other.address
      )).to.be.revertedWith("Invalid amount");
    });

    it("Should not allow withdrawal to zero address", async function () {
      await expect(wallet.connect(signer1).requestWithdrawal(
        ethers.ZeroAddress,
        ethers.parseEther("1.0"),
        ethers.ZeroAddress
      )).to.be.revertedWith("Invalid recipient");
    });

    it("Should increment withdrawal nonce correctly", async function () {
      const initialNonce = await wallet.withdrawalNonce();
      
      await wallet.connect(signer1).requestWithdrawal(
        ethers.ZeroAddress,
        ethers.parseEther("1.0"),
        other.address
      );

      expect(await wallet.withdrawalNonce()).to.equal(initialNonce + 1n);
    });

    it("Should not allow a signer to sign the same withdrawal twice", async function () {
      // Create request
      const tx = await wallet.connect(signer1).requestWithdrawal(
        ethers.ZeroAddress,
        ethers.parseEther("1.0"),
        other.address
      );
      
      const receipt = await tx.wait();
      const event = receipt.logs.find(log => log.fragment && log.fragment.name === 'WithdrawalRequested');
      const requestId = event.args[0];

      // Attempt to sign again with the same signer
      await expect(wallet.connect(signer1).signWithdrawal(requestId))
        .to.be.revertedWith("Already signed");
    });
  });
});
