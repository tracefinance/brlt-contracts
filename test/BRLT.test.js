const { ethers } = require("hardhat");
const { expect } = require("chai");

const { loadFixture } = require("@nomicfoundation/hardhat-network-helpers");
const { upgrades } = require("hardhat");

// Helper function to convert amounts to the token's decimal representation
function H(amount) {
    return ethers.parseUnits(amount.toString(), 18); // BRLT has 18 decimals
}

describe("BRLT", function () {
    // We define a fixture to reuse the same setup in every test.
    // We use loadFixture to run this setup once, snapshot that state,
    // and reset Hardhat Network to that snapshot in every test.
    async function deployBRLTFixture() {
        // Contracts are deployed using the first signer/account by default
        const [owner, user1, user2, otherAccount] = await ethers.getSigners();

        const BRLT = await ethers.getContractFactory("BRLT", owner);
        // Deploy as an upgradeable proxy
        const brlt = await upgrades.deployProxy(BRLT, [owner.address], { 
            kind: 'uups', 
            initializer: 'initialize'
        });
        await brlt.waitForDeployment();

        // Roles for convenience (matches BRLT.sol)
        const MINTER_ROLE = await brlt.MINTER_ROLE();
        const BURNER_ROLE = await brlt.BURNER_ROLE();
        const PAUSER_ROLE = await brlt.PAUSER_ROLE();
        const BLACKLISTER_ROLE = await brlt.BLACKLISTER_ROLE();
        const DEFAULT_ADMIN_ROLE = await brlt.DEFAULT_ADMIN_ROLE();
        const UPGRADER_ROLE = await brlt.UPGRADER_ROLE();

        return { 
            brlt, 
            owner, user1, user2, otherAccount, 
            MINTER_ROLE, BURNER_ROLE, PAUSER_ROLE, BLACKLISTER_ROLE, DEFAULT_ADMIN_ROLE, UPGRADER_ROLE 
        };
    }

    // Helper function for EIP-2612 permit signatures
    async function getPermitSignature(signer, token, spenderAddress, value, deadline, nonce) {
        const domain = {
            name: await token.name(),
            version: "1", // Default version for ERC20Permit
            chainId: (await ethers.provider.getNetwork()).chainId,
            verifyingContract: await token.getAddress()
        };

        const types = {
            Permit: [
                { name: "owner", type: "address" },
                { name: "spender", type: "address" },
                { name: "value", type: "uint256" },
                { name: "nonce", type: "uint256" },
                { name: "deadline", type: "uint256" }
            ]
        };

        const message = {
            owner: signer.address,
            spender: spenderAddress,
            value: value,
            nonce: nonce || (await token.nonces(signer.address)),
            deadline: deadline || Math.floor(Date.now() / 1000) + 3600 // 1 hour from now
        };

        const signature = await signer.signTypedData(domain, types, message);
        return { ...message, signature };
    }

    describe("Deployment", function () {
        it("Should set the right token name and symbol", async function () {
            const { brlt } = await loadFixture(deployBRLTFixture);
            expect(await brlt.name()).to.equal("BRLT");
            expect(await brlt.symbol()).to.equal("BRLT");
        });

        it("Should have 18 decimals", async function () {
            const { brlt } = await loadFixture(deployBRLTFixture);
            expect(await brlt.decimals()).to.equal(18);
        });

        it("Should assign all roles (MINTER, BURNER, PAUSER, BLACKLISTER, DEFAULT_ADMIN, UPGRADER) to the deployer (initialAdmin)", async function () {
            const { brlt, owner, MINTER_ROLE, BURNER_ROLE, PAUSER_ROLE, BLACKLISTER_ROLE, DEFAULT_ADMIN_ROLE, UPGRADER_ROLE } = await loadFixture(deployBRLTFixture);
            expect(await brlt.hasRole(DEFAULT_ADMIN_ROLE, owner.address)).to.be.true;
            expect(await brlt.hasRole(MINTER_ROLE, owner.address)).to.be.true;
            expect(await brlt.hasRole(BURNER_ROLE, owner.address)).to.be.true;
            expect(await brlt.hasRole(PAUSER_ROLE, owner.address)).to.be.true;
            expect(await brlt.hasRole(BLACKLISTER_ROLE, owner.address)).to.be.true;
            expect(await brlt.hasRole(UPGRADER_ROLE, owner.address)).to.be.true;
        });

        it("Should have an initial total supply of 0", async function () {
            const { brlt } = await loadFixture(deployBRLTFixture);
            expect(await brlt.totalSupply()).to.equal(0);
        });

        it("Should support IAccessControl interface (0x7965db0b) via super", async function () {
            const { brlt } = await loadFixture(deployBRLTFixture);
            const IAccessControlInterfaceId = "0x7965db0b";
            expect(await brlt.supportsInterface(IAccessControlInterfaceId)).to.be.true;
        });

        it("Should support IERC20Permit interface (0xd505accf)", async function () {
            const { brlt } = await loadFixture(deployBRLTFixture);
            const IERC20PermitInterfaceId = "0xd505accf"; // As defined in BRLT.sol
            expect(await brlt.supportsInterface(IERC20PermitInterfaceId)).to.be.true;
        });
    });

    describe("ERC20 Standard Functionality", function () {
        const mintAmount = H(1000); // Mint 1000 BRLT for tests

        it("Should correctly report balanceOf after minting", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(user1.address, mintAmount);
            expect(await brlt.balanceOf(user1.address)).to.equal(mintAmount);
            expect(await brlt.balanceOf(owner.address)).to.equal(0); // Owner didn't get tokens directly
        });

        it("Should transfer tokens between accounts", async function () {
            const { brlt, owner, user1, user2 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(user1.address, mintAmount);

            const transferAmount = H(100);
            await expect(brlt.connect(user1).transfer(user2.address, transferAmount))
                .to.emit(brlt, "Transfer")
                .withArgs(user1.address, user2.address, transferAmount);

            expect(await brlt.balanceOf(user1.address)).to.equal(mintAmount - transferAmount);
            expect(await brlt.balanceOf(user2.address)).to.equal(transferAmount);
        });

        it("Should fail to transfer if balance is insufficient", async function () {
            const { brlt, owner, user1, user2 } = await loadFixture(deployBRLTFixture);
            // user1 has 0 tokens initially
            const transferAmount = H(100);
            await expect(brlt.connect(user1).transfer(user2.address, transferAmount))
                .to.be.revertedWithCustomError(brlt, "ERC20InsufficientBalance");
        });

        it("Should fail to transfer to the zero address", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(user1.address, mintAmount);
            const transferAmount = H(100);
            await expect(brlt.connect(user1).transfer(ethers.ZeroAddress, transferAmount))
                .to.be.revertedWithCustomError(brlt, "ERC20InvalidReceiver").withArgs(ethers.ZeroAddress);
        });

        it("Should approve spender and update allowance", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            const approveAmount = H(500);

            await expect(brlt.connect(owner).approve(user1.address, approveAmount))
                .to.emit(brlt, "Approval")
                .withArgs(owner.address, user1.address, approveAmount);
            
            expect(await brlt.allowance(owner.address, user1.address)).to.equal(approveAmount);
        });

        it("Should allow spender to transferFrom if approved", async function () {
            const { brlt, owner, user1, user2 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(owner.address, mintAmount); // Mint to owner

            const approveAmount = H(500);
            await brlt.connect(owner).approve(user1.address, approveAmount);

            const transferAmount = H(300);
            await expect(brlt.connect(user1).transferFrom(owner.address, user2.address, transferAmount))
                .to.emit(brlt, "Transfer")
                .withArgs(owner.address, user2.address, transferAmount);

            expect(await brlt.balanceOf(owner.address)).to.equal(mintAmount - transferAmount);
            expect(await brlt.balanceOf(user2.address)).to.equal(transferAmount);
            expect(await brlt.allowance(owner.address, user1.address)).to.equal(approveAmount - transferAmount);
        });

        it("Should fail transferFrom if spender is not approved", async function () {
            const { brlt, owner, user1, user2 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(owner.address, mintAmount);

            const transferAmount = H(100);
            await expect(brlt.connect(user1).transferFrom(owner.address, user2.address, transferAmount))
                .to.be.revertedWithCustomError(brlt, "ERC20InsufficientAllowance").withArgs(user1.address, 0, transferAmount); 
        });

        it("Should fail transferFrom if approved amount is insufficient", async function () {
            const { brlt, owner, user1, user2 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(owner.address, mintAmount);

            const approveAmount = H(50);
            await brlt.connect(owner).approve(user1.address, approveAmount);

            const transferAmount = H(100);
            await expect(brlt.connect(user1).transferFrom(owner.address, user2.address, transferAmount))
                .to.be.revertedWithCustomError(brlt, "ERC20InsufficientAllowance").withArgs(user1.address, approveAmount, transferAmount);
        });

        it("Should handle approval to the zero address (typically allowed, sets allowance to 0)", async function () {
            const { brlt, owner } = await loadFixture(deployBRLTFixture);
            const approveAmount = H(100);
            // Standard ERC20 allows approving the zero address, though it has no practical effect for spending.
            // Some implementations might revert, OpenZeppelin's does not by default for `approve`.
            // UPDATE: OpenZeppelin v5.x ERC20.sol *does* revert this.
            await expect(brlt.connect(owner).approve(ethers.ZeroAddress, approveAmount))
                .to.be.revertedWithCustomError(brlt, "ERC20InvalidSpender").withArgs(ethers.ZeroAddress);
            // expect(await brlt.allowance(owner.address, ethers.ZeroAddress)).to.equal(approveAmount); // This line is no longer reachable
        });
    });

    describe("Minting", function () {
        const mintAmount = H(100);

        it("Should allow MINTER_ROLE to mint tokens", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            await expect(brlt.connect(owner).mint(user1.address, mintAmount))
                .to.emit(brlt, "Transfer")
                .withArgs(ethers.ZeroAddress, user1.address, mintAmount);
            expect(await brlt.balanceOf(user1.address)).to.equal(mintAmount);
            expect(await brlt.totalSupply()).to.equal(mintAmount);
        });

        it("Should NOT allow an account without MINTER_ROLE to mint tokens", async function () {
            const { brlt, user1, user2, MINTER_ROLE } = await loadFixture(deployBRLTFixture);
            await expect(brlt.connect(user1).mint(user2.address, mintAmount))
                .to.be.revertedWithCustomError(brlt, "AccessControlUnauthorizedAccount")
                .withArgs(user1.address, MINTER_ROLE);
        });

        it("Should NOT allow minting to the zero address", async function () {
            const { brlt, owner } = await loadFixture(deployBRLTFixture);
            await expect(brlt.connect(owner).mint(ethers.ZeroAddress, mintAmount))
                .to.be.revertedWithCustomError(brlt, "ERC20InvalidReceiver").withArgs(ethers.ZeroAddress);
        });

        it("Should NOT allow minting when paused", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).pause();
            await expect(brlt.connect(owner).mint(user1.address, mintAmount))
                .to.be.revertedWithCustomError(brlt, "EnforcedPause");
        });

        it("Should NOT allow minting to a blacklisted address", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).blacklistAddress(user1.address);
            await expect(brlt.connect(owner).mint(user1.address, mintAmount))
                .to.be.revertedWithCustomError(brlt, "AccountBlacklisted").withArgs(user1.address);
        });

        it("Should increase totalSupply on mint", async function () {
            const { brlt, owner, user1, user2 } = await loadFixture(deployBRLTFixture);
            const initialSupply = await brlt.totalSupply();
            expect(initialSupply).to.equal(0);

            await brlt.connect(owner).mint(user1.address, mintAmount);
            expect(await brlt.totalSupply()).to.equal(initialSupply + mintAmount);

            const anotherMintAmount = H(50);
            await brlt.connect(owner).mint(user2.address, anotherMintAmount);
            expect(await brlt.totalSupply()).to.equal(initialSupply + mintAmount + anotherMintAmount);
        });
    });

    describe("Burning (burnFrom)", function () {
        const mintAmount = H(1000);
        const burnAmount = H(100);

        it("Should allow BURNER_ROLE to burn tokens from an account", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(user1.address, mintAmount);
            
            await expect(brlt.connect(owner).burnFrom(user1.address, burnAmount))
                .to.emit(brlt, "Transfer")
                .withArgs(user1.address, ethers.ZeroAddress, burnAmount);
            
            expect(await brlt.balanceOf(user1.address)).to.equal(mintAmount - burnAmount);
            expect(await brlt.totalSupply()).to.equal(mintAmount - burnAmount);
        });

        it("Should allow BURNER_ROLE to burn their own tokens", async function () {
            const { brlt, owner } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(owner.address, mintAmount); // Mint to owner (burner)
            
            await expect(brlt.connect(owner).burnFrom(owner.address, burnAmount))
                .to.emit(brlt, "Transfer")
                .withArgs(owner.address, ethers.ZeroAddress, burnAmount);
            
            expect(await brlt.balanceOf(owner.address)).to.equal(mintAmount - burnAmount);
            expect(await brlt.totalSupply()).to.equal(mintAmount - burnAmount);
        });

        it("Should NOT allow an account without BURNER_ROLE to burn tokens", async function () {
            const { brlt, owner, user1, BURNER_ROLE } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(user1.address, mintAmount);

            await expect(brlt.connect(user1).burnFrom(user1.address, burnAmount))
                .to.be.revertedWithCustomError(brlt, "AccessControlUnauthorizedAccount")
                .withArgs(user1.address, BURNER_ROLE);
        });

        it("Should NOT allow burning more tokens than an account has", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(user1.address, mintAmount); // user1 has 1000
            const excessiveBurnAmount = H(1001);

            await expect(brlt.connect(owner).burnFrom(user1.address, excessiveBurnAmount))
                .to.be.revertedWithCustomError(brlt, "ERC20InsufficientBalance")
                .withArgs(user1.address, mintAmount, excessiveBurnAmount);
        });

        it("Should NOT allow burning from the zero address", async function () {
            const { brlt, owner } = await loadFixture(deployBRLTFixture);
            // Attempting to burn from zero address doesn't make sense as it can't own tokens.
            // _burn function in OZ ERC20.sol requires account != address(0) if `from` is the account.
            // Our burnFrom explicitly passes an account.
            await expect(brlt.connect(owner).burnFrom(ethers.ZeroAddress, burnAmount))
                 .to.be.revertedWithCustomError(brlt, "ERC20InvalidSender") // OpenZeppelin's _burn check
                 .withArgs(ethers.ZeroAddress);
        });
        
        it("Should NOT allow burning when paused", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(user1.address, mintAmount);
            await brlt.connect(owner).pause();

            await expect(brlt.connect(owner).burnFrom(user1.address, burnAmount))
                .to.be.revertedWithCustomError(brlt, "EnforcedPause");
        });

        it("Should NOT allow burning from a blacklisted account (sender of funds)", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(user1.address, mintAmount);
            await brlt.connect(owner).blacklistAddress(user1.address);

            await expect(brlt.connect(owner).burnFrom(user1.address, burnAmount))
                .to.be.revertedWithCustomError(brlt, "AccountBlacklisted")
                .withArgs(user1.address);
        });

        it("Should allow burning if the burner is blacklisted but the source account is not", async function () {
            const { brlt, owner, user1, user2, BURNER_ROLE } = await loadFixture(deployBRLTFixture);
            // Grant user2 BURNER_ROLE
            await brlt.connect(owner).grantRole(BURNER_ROLE, user2.address);
            // Mint tokens to user1
            await brlt.connect(owner).mint(user1.address, mintAmount);
            // Blacklist user2 (the burner)
            await brlt.connect(owner).blacklistAddress(user2.address);

            // User2 (burner, blacklisted) tries to burn from user1 (not blacklisted)
            // This should be allowed because the _update check is on `from` (user1) and `to` (address(0))
            // not the msg.sender (user2) of the burnFrom call.
            await expect(brlt.connect(user2).burnFrom(user1.address, burnAmount))
                .to.emit(brlt, "Transfer")
                .withArgs(user1.address, ethers.ZeroAddress, burnAmount);

            expect(await brlt.balanceOf(user1.address)).to.equal(mintAmount - burnAmount);
        });
    });

    describe("Pausable Functionality", function () {
        const mintAmount = H(100);
        const transferAmount = H(50);
        const burnAmount = H(50);

        it("Should be unpaused by default", async function () {
            const { brlt } = await loadFixture(deployBRLTFixture);
            expect(await brlt.paused()).to.be.false;
        });

        it("Should allow PAUSER_ROLE to pause and unpause the contract", async function () {
            const { brlt, owner } = await loadFixture(deployBRLTFixture);
            
            // Pause
            await expect(brlt.connect(owner).pause())
                .to.emit(brlt, "Paused")
                .withArgs(owner.address);
            expect(await brlt.paused()).to.be.true;

            // Unpause
            await expect(brlt.connect(owner).unpause())
                .to.emit(brlt, "Unpaused")
                .withArgs(owner.address);
            expect(await brlt.paused()).to.be.false;
        });

        it("Should NOT allow an account without PAUSER_ROLE to pause", async function () {
            const { brlt, user1, PAUSER_ROLE } = await loadFixture(deployBRLTFixture);
            await expect(brlt.connect(user1).pause())
                .to.be.revertedWithCustomError(brlt, "AccessControlUnauthorizedAccount")
                .withArgs(user1.address, PAUSER_ROLE);
        });

        it("Should NOT allow an account without PAUSER_ROLE to unpause", async function () {
            const { brlt, owner, user1, PAUSER_ROLE } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).pause(); // Paused by owner
            await expect(brlt.connect(user1).unpause())
                .to.be.revertedWithCustomError(brlt, "AccessControlUnauthorizedAccount")
                .withArgs(user1.address, PAUSER_ROLE);
        });

        it("Should prevent transfers when paused", async function () {
            const { brlt, owner, user1, user2 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(user1.address, mintAmount);
            await brlt.connect(owner).pause();
            await expect(brlt.connect(user1).transfer(user2.address, transferAmount))
                .to.be.revertedWithCustomError(brlt, "EnforcedPause");
        });

        it("Should prevent minting when paused", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).pause();
            await expect(brlt.connect(owner).mint(user1.address, mintAmount))
                .to.be.revertedWithCustomError(brlt, "EnforcedPause");
        });

        it("Should prevent burning (burnFrom) when paused", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(user1.address, mintAmount);
            await brlt.connect(owner).pause();
            await expect(brlt.connect(owner).burnFrom(user1.address, burnAmount))
                .to.be.revertedWithCustomError(brlt, "EnforcedPause");
        });

        it("Should allow operations after unpausing", async function () {
            const { brlt, owner, user1, user2 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(user1.address, mintAmount);
            
            // Pause and unpause
            await brlt.connect(owner).pause();
            await brlt.connect(owner).unpause();

            // Transfer should now work
            await expect(brlt.connect(user1).transfer(user2.address, transferAmount))
                .to.emit(brlt, "Transfer")
                .withArgs(user1.address, user2.address, transferAmount);
            expect(await brlt.balanceOf(user2.address)).to.equal(transferAmount);

            // Mint should now work
            const newMintAmount = H(50);
            await expect(brlt.connect(owner).mint(user1.address, newMintAmount))
                .to.emit(brlt, "Transfer");
            // user1 had mintAmount (100) - transferAmount (50) = 50. Now + 50 = 100
            expect(await brlt.balanceOf(user1.address)).to.equal(mintAmount - transferAmount + newMintAmount);

            // Burn should now work
            // user1 has 100. Burn 50.
            await expect(brlt.connect(owner).burnFrom(user1.address, burnAmount))
                .to.emit(brlt, "Transfer");
            expect(await brlt.balanceOf(user1.address)).to.equal(mintAmount - transferAmount + newMintAmount - burnAmount);
        });

        it("Should revert when trying to pause if already paused", async function () {
            const { brlt, owner } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).pause();
            await expect(brlt.connect(owner).pause())
                .to.be.revertedWithCustomError(brlt, "EnforcedPause");
        });

        it("Should revert when trying to unpause if not paused", async function () {
            const { brlt, owner } = await loadFixture(deployBRLTFixture);
            // Already unpaused by default
            await expect(brlt.connect(owner).unpause())
                .to.be.revertedWithCustomError(brlt, "ExpectedPause");
        });
    });

    describe("Blacklisting Functionality", function () {
        const mintAmount = H(1000);
        const transferAmount = H(100);

        it("BLACKLISTER_ROLE should be able to blacklist and unblacklist an address", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            expect(await brlt.isBlacklisted(user1.address)).to.be.false;

            // Blacklist
            await expect(brlt.connect(owner).blacklistAddress(user1.address))
                .to.emit(brlt, "AddressBlacklisted")
                .withArgs(user1.address);
            expect(await brlt.isBlacklisted(user1.address)).to.be.true;

            // Unblacklist
            await expect(brlt.connect(owner).unblacklistAddress(user1.address))
                .to.emit(brlt, "AddressUnblacklisted")
                .withArgs(user1.address);
            expect(await brlt.isBlacklisted(user1.address)).to.be.false;
        });

        it("Should NOT allow an account without BLACKLISTER_ROLE to blacklist", async function () {
            const { brlt, user1, user2, BLACKLISTER_ROLE } = await loadFixture(deployBRLTFixture);
            await expect(brlt.connect(user1).blacklistAddress(user2.address))
                .to.be.revertedWithCustomError(brlt, "AccessControlUnauthorizedAccount")
                .withArgs(user1.address, BLACKLISTER_ROLE);
        });

        it("Should NOT allow an account without BLACKLISTER_ROLE to unblacklist", async function () {
            const { brlt, owner, user1, user2, BLACKLISTER_ROLE } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).blacklistAddress(user2.address); // Blacklist first
            await expect(brlt.connect(user1).unblacklistAddress(user2.address))
                .to.be.revertedWithCustomError(brlt, "AccessControlUnauthorizedAccount")
                .withArgs(user1.address, BLACKLISTER_ROLE);
        });

        it("Should NOT allow blacklisting the zero address", async function () {
            const { brlt, owner } = await loadFixture(deployBRLTFixture);
            await expect(brlt.connect(owner).blacklistAddress(ethers.ZeroAddress))
                .to.be.revertedWith("BRLT: cannot blacklist zero address");
        });

        it("Should prevent a blacklisted account from sending tokens (transfer)", async function () {
            const { brlt, owner, user1, user2 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(user1.address, mintAmount);
            await brlt.connect(owner).blacklistAddress(user1.address);

            await expect(brlt.connect(user1).transfer(user2.address, transferAmount))
                .to.be.revertedWithCustomError(brlt, "AccountBlacklisted")
                .withArgs(user1.address);
        });

        it("Should prevent sending tokens to a blacklisted account (transfer)", async function () {
            const { brlt, owner, user1, user2 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(user1.address, mintAmount);
            await brlt.connect(owner).blacklistAddress(user2.address);

            await expect(brlt.connect(user1).transfer(user2.address, transferAmount))
                .to.be.revertedWithCustomError(brlt, "AccountBlacklisted")
                .withArgs(user2.address);
        });

        it("Should prevent minting to a blacklisted account", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).blacklistAddress(user1.address);
            await expect(brlt.connect(owner).mint(user1.address, mintAmount))
                .to.be.revertedWithCustomError(brlt, "AccountBlacklisted")
                .withArgs(user1.address);
        });

        // Note: Burning from a blacklisted account is already tested in "Burning (burnFrom)" section.
        // It correctly reverts with AccountBlacklisted(user1.address).

        it("Should allow operations once an account is unblacklisted", async function () {
            const { brlt, owner, user1, user2 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(user1.address, mintAmount);
            
            // Blacklist and then unblacklist user1
            await brlt.connect(owner).blacklistAddress(user1.address);
            await brlt.connect(owner).unblacklistAddress(user1.address);

            // Transfer should now work
            await expect(brlt.connect(user1).transfer(user2.address, transferAmount))
                .to.emit(brlt, "Transfer")
                .withArgs(user1.address, user2.address, transferAmount);
            expect(await brlt.balanceOf(user2.address)).to.equal(transferAmount);
        });

        it("Blacklisting should not affect non-blacklisted accounts", async function () {
            const { brlt, owner, user1, user2, otherAccount } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(user1.address, mintAmount);
            await brlt.connect(owner).mint(user2.address, mintAmount);

            // Blacklist otherAccount
            await brlt.connect(owner).blacklistAddress(otherAccount.address);

            // Transfer between user1 and user2 should still work
            await expect(brlt.connect(user1).transfer(user2.address, transferAmount))
                .to.emit(brlt, "Transfer");
            expect(await brlt.balanceOf(user1.address)).to.equal(mintAmount - transferAmount);
            expect(await brlt.balanceOf(user2.address)).to.equal(mintAmount + transferAmount);
        });
    });

    describe("EIP-2612 Permit Functionality", function () {
        const mintAmount = H(1000);
        const approveAmount = H(500);

        it("Should allow spending with a valid permit", async function () {
            const { brlt, owner, user1, user2 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(owner.address, mintAmount); // owner has tokens

            const deadline = Math.floor(Date.now() / 1000) + 3600;
            const nonce = await brlt.nonces(owner.address);
            const {v, r, s} = ethers.Signature.from( (await getPermitSignature(owner, brlt, user1.address, approveAmount, deadline, nonce)).signature );

            await expect(brlt.permit(owner.address, user1.address, approveAmount, deadline, v, r, s))
                .to.emit(brlt, "Approval")
                .withArgs(owner.address, user1.address, approveAmount);

            expect(await brlt.allowance(owner.address, user1.address)).to.equal(approveAmount);
            expect(await brlt.nonces(owner.address)).to.equal(nonce + 1n); // Nonce should increment

            // user1 can now spend tokens
            const transferAmount = H(100);
            await expect(brlt.connect(user1).transferFrom(owner.address, user2.address, transferAmount))
                .to.emit(brlt, "Transfer");
            expect(await brlt.balanceOf(user2.address)).to.equal(transferAmount);
        });

        it("Should reject expired permit", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            const deadline = Math.floor(Date.now() / 1000) - 3600; // 1 hour in the past
            const nonce = await brlt.nonces(owner.address);
            const {v, r, s} = ethers.Signature.from( (await getPermitSignature(owner, brlt, user1.address, approveAmount, deadline, nonce)).signature );

            await expect(brlt.permit(owner.address, user1.address, approveAmount, deadline, v, r, s))
                .to.be.revertedWithCustomError(brlt, "ERC2612ExpiredSignature");
        });

        it("Should reject permit with invalid signature (wrong signer)", async function () {
            const { brlt, owner, user1, otherAccount } = await loadFixture(deployBRLTFixture);
            const deadline = Math.floor(Date.now() / 1000) + 3600;
            const nonce = await brlt.nonces(owner.address);
            // Signature by otherAccount, but owner is owner.address
            const {v, r, s} = ethers.Signature.from( (await getPermitSignature(otherAccount, brlt, user1.address, approveAmount, deadline, nonce)).signature );

            await expect(brlt.permit(owner.address, user1.address, approveAmount, deadline, v, r, s))
                .to.be.revertedWithCustomError(brlt, "ERC2612InvalidSigner");
        });

        it("Should reject permit with used nonce (replay attack)", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            const deadline = Math.floor(Date.now() / 1000) + 3600;
            const nonce = await brlt.nonces(owner.address);
            const {v, r, s} = ethers.Signature.from( (await getPermitSignature(owner, brlt, user1.address, approveAmount, deadline, nonce)).signature );

            // Use the permit once (should succeed)
            await brlt.permit(owner.address, user1.address, approveAmount, deadline, v, r, s);
            expect(await brlt.nonces(owner.address)).to.equal(nonce + 1n);

            // Try to use the same permit again (should fail due to nonce)
            await expect(brlt.permit(owner.address, user1.address, approveAmount, deadline, v, r, s))
                .to.be.revertedWithCustomError(brlt, "ERC2612InvalidSigner"); // Nonce is part of signed message, so changing it invalidates sig
        });

        it("Should reject permit when contract is paused", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            const deadline = Math.floor(Date.now() / 1000) + 3600;
            const nonce = await brlt.nonces(owner.address);
            const {v, r, s} = ethers.Signature.from( (await getPermitSignature(owner, brlt, user1.address, approveAmount, deadline, nonce)).signature );

            await brlt.connect(owner).pause();
            await expect(brlt.permit(owner.address, user1.address, approveAmount, deadline, v, r, s))
                .to.be.revertedWithCustomError(brlt, "EnforcedPause");
        });

        // ERC20Permit's permit function itself doesn't directly interact with _update (which has blacklist checks).
        // The blacklist check will occur when transferFrom is attempted using the allowance granted by permit.
        // So, we test that transferFrom fails if owner or spender is blacklisted after permit.

        it("transferFrom using permit allowance should fail if owner is blacklisted after permit", async function () {
            const { brlt, owner, user1, user2 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(owner.address, mintAmount);
            const deadline = Math.floor(Date.now() / 1000) + 3600;
            const nonce = await brlt.nonces(owner.address);
            const {v, r, s} = ethers.Signature.from( (await getPermitSignature(owner, brlt, user1.address, approveAmount, deadline, nonce)).signature );

            await brlt.permit(owner.address, user1.address, approveAmount, deadline, v, r, s);
            await brlt.connect(owner).blacklistAddress(owner.address); // Blacklist owner AFTER permit

            const transferAmount = H(100);
            await expect(brlt.connect(user1).transferFrom(owner.address, user2.address, transferAmount))
                .to.be.revertedWithCustomError(brlt, "AccountBlacklisted")
                .withArgs(owner.address);
        });

        it("transferFrom using permit allowance should fail if spender (receiver in transferFrom) is blacklisted after permit", async function () {
            const { brlt, owner, user1, user2 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(owner.address, mintAmount);
            const deadline = Math.floor(Date.now() / 1000) + 3600;
            const nonce = await brlt.nonces(owner.address);
            const {v, r, s} = ethers.Signature.from( (await getPermitSignature(owner, brlt, user1.address, approveAmount, deadline, nonce)).signature );

            await brlt.permit(owner.address, user1.address, approveAmount, deadline, v, r, s);
            await brlt.connect(owner).blacklistAddress(user2.address); // Blacklist receiver (user2) AFTER permit

            const transferAmount = H(100);
            await expect(brlt.connect(user1).transferFrom(owner.address, user2.address, transferAmount))
                .to.be.revertedWithCustomError(brlt, "AccountBlacklisted")
                .withArgs(user2.address);
        });

        it("Should allow permit with zero value (sets allowance to 0, emits event)", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(owner.address, mintAmount); // owner has tokens

            // First, set some allowance via permit
            const deadline = Math.floor(Date.now() / 1000) + 3600;
            let nonce = await brlt.nonces(owner.address);
            const {v: v1, r: r1, s: s1} = ethers.Signature.from( (await getPermitSignature(owner, brlt, user1.address, H(100), deadline, nonce)).signature );
            await brlt.permit(owner.address, user1.address, H(100), deadline, v1, r1, s1);
            expect(await brlt.allowance(owner.address, user1.address)).to.equal(H(100));
            nonce = await brlt.nonces(owner.address); // nonce has incremented

            // Then, permit zero value
            const {v: v2, r: r2, s: s2} = ethers.Signature.from( (await getPermitSignature(owner, brlt, user1.address, 0, deadline, nonce)).signature );
            await expect(brlt.permit(owner.address, user1.address, 0, deadline, v2, r2, s2))
                .to.emit(brlt, "Approval")
                .withArgs(owner.address, user1.address, 0);
            
            expect(await brlt.allowance(owner.address, user1.address)).to.equal(0);
            expect(await brlt.nonces(owner.address)).to.equal(nonce + 1n);
        });

    });

    describe("Access Control", function () {
        it("DEFAULT_ADMIN_ROLE (owner) should be able to grant roles", async function () {
            const { brlt, owner, user1, MINTER_ROLE } = await loadFixture(deployBRLTFixture);
            expect(await brlt.hasRole(MINTER_ROLE, user1.address)).to.be.false;
            
            await expect(brlt.connect(owner).grantRole(MINTER_ROLE, user1.address))
                .to.emit(brlt, "RoleGranted")
                .withArgs(MINTER_ROLE, user1.address, owner.address);
            
            expect(await brlt.hasRole(MINTER_ROLE, user1.address)).to.be.true;
        });

        it("DEFAULT_ADMIN_ROLE (owner) should be able to revoke roles", async function () {
            const { brlt, owner, user1, MINTER_ROLE } = await loadFixture(deployBRLTFixture);
            // Grant role first
            await brlt.connect(owner).grantRole(MINTER_ROLE, user1.address);
            expect(await brlt.hasRole(MINTER_ROLE, user1.address)).to.be.true;

            // Revoke role
            await expect(brlt.connect(owner).revokeRole(MINTER_ROLE, user1.address))
                .to.emit(brlt, "RoleRevoked")
                .withArgs(MINTER_ROLE, user1.address, owner.address);
            
            expect(await brlt.hasRole(MINTER_ROLE, user1.address)).to.be.false;
        });

        it("An account without DEFAULT_ADMIN_ROLE should NOT be able to grant roles", async function () {
            const { brlt, user1, user2, MINTER_ROLE, DEFAULT_ADMIN_ROLE } = await loadFixture(deployBRLTFixture);
            // user1 does not have DEFAULT_ADMIN_ROLE
            await expect(brlt.connect(user1).grantRole(MINTER_ROLE, user2.address))
                .to.be.revertedWithCustomError(brlt, "AccessControlUnauthorizedAccount")
                .withArgs(user1.address, DEFAULT_ADMIN_ROLE);
        });

        it("An account without DEFAULT_ADMIN_ROLE should NOT be able to revoke roles", async function () {
            const { brlt, owner, user1, user2, MINTER_ROLE, DEFAULT_ADMIN_ROLE } = await loadFixture(deployBRLTFixture);
            // Grant MINTER_ROLE to user2 by owner first
            await brlt.connect(owner).grantRole(MINTER_ROLE, user2.address);
            expect(await brlt.hasRole(MINTER_ROLE, user2.address)).to.be.true;

            // user1 (not admin) tries to revoke
            await expect(brlt.connect(user1).revokeRole(MINTER_ROLE, user2.address))
                .to.be.revertedWithCustomError(brlt, "AccessControlUnauthorizedAccount")
                .withArgs(user1.address, DEFAULT_ADMIN_ROLE);
        });

        it("Account should be able to renounce its own role", async function () {
            const { brlt, owner, user1, MINTER_ROLE } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).grantRole(MINTER_ROLE, user1.address);
            expect(await brlt.hasRole(MINTER_ROLE, user1.address)).to.be.true;

            // user1 renounces their MINTER_ROLE
            await expect(brlt.connect(user1).renounceRole(MINTER_ROLE, user1.address))
                .to.emit(brlt, "RoleRevoked")
                .withArgs(MINTER_ROLE, user1.address, user1.address);
            
            expect(await brlt.hasRole(MINTER_ROLE, user1.address)).to.be.false;
        });

        it("Renouncing a role not held should not revert (AccessControl behavior)", async function () {
            const { brlt, user1, MINTER_ROLE } = await loadFixture(deployBRLTFixture);
            expect(await brlt.hasRole(MINTER_ROLE, user1.address)).to.be.false;
            // User1 tries to renounce MINTER_ROLE which they don't have.
            // OpenZeppelin's AccessControl _revokeRole doesn't revert if role not held,
            // and it also does NOT emit RoleRevoked in this case.
            await expect(brlt.connect(user1).renounceRole(MINTER_ROLE, user1.address))
                .to.not.emit(brlt, "RoleRevoked"); // Adjusted expectation
            expect(await brlt.hasRole(MINTER_ROLE, user1.address)).to.be.false;
        });

        it("Renouncing DEFAULT_ADMIN_ROLE should be possible", async function () {
            const { brlt, owner, DEFAULT_ADMIN_ROLE } = await loadFixture(deployBRLTFixture);
            expect(await brlt.hasRole(DEFAULT_ADMIN_ROLE, owner.address)).to.be.true;
            
            await expect(brlt.connect(owner).renounceRole(DEFAULT_ADMIN_ROLE, owner.address))
                .to.emit(brlt, "RoleRevoked")
                .withArgs(DEFAULT_ADMIN_ROLE, owner.address, owner.address);
            
            expect(await brlt.hasRole(DEFAULT_ADMIN_ROLE, owner.address)).to.be.false;
        });

        it("Once DEFAULT_ADMIN_ROLE is renounced, admin functions are no longer possible by that account", async function () {
            const { brlt, owner, user1, MINTER_ROLE, DEFAULT_ADMIN_ROLE } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).renounceRole(DEFAULT_ADMIN_ROLE, owner.address);
            expect(await brlt.hasRole(DEFAULT_ADMIN_ROLE, owner.address)).to.be.false;

            // Owner (ex-admin) tries to grant a role
            await expect(brlt.connect(owner).grantRole(MINTER_ROLE, user1.address))
                .to.be.revertedWithCustomError(brlt, "AccessControlUnauthorizedAccount")
                .withArgs(owner.address, DEFAULT_ADMIN_ROLE);
        });
    });

    describe("Edge Cases and Other Scenarios", function () {
        const mintAmount = H(100);

        it("Should allow minting zero tokens (no state change, emits event)", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            const initialTotalSupply = await brlt.totalSupply();
            const initialBalance = await brlt.balanceOf(user1.address);

            await expect(brlt.connect(owner).mint(user1.address, 0))
                .to.emit(brlt, "Transfer")
                .withArgs(ethers.ZeroAddress, user1.address, 0);

            expect(await brlt.totalSupply()).to.equal(initialTotalSupply);
            expect(await brlt.balanceOf(user1.address)).to.equal(initialBalance);
        });

        it("Should allow burning zero tokens from an account (no state change, emits event)", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(user1.address, mintAmount); 

            const initialTotalSupply = await brlt.totalSupply();
            const initialBalance = await brlt.balanceOf(user1.address);

            await expect(brlt.connect(owner).burnFrom(user1.address, 0))
                .to.emit(brlt, "Transfer")
                .withArgs(user1.address, ethers.ZeroAddress, 0);

            expect(await brlt.totalSupply()).to.equal(initialTotalSupply);
            expect(await brlt.balanceOf(user1.address)).to.equal(initialBalance);
        });

        it("Should allow approving zero tokens (sets allowance to 0, emits event)", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            // First, approve some amount
            await brlt.connect(owner).approve(user1.address, H(100));
            expect(await brlt.allowance(owner.address, user1.address)).to.equal(H(100));

            // Then approve zero
            await expect(brlt.connect(owner).approve(user1.address, 0))
                .to.emit(brlt, "Approval")
                .withArgs(owner.address, user1.address, 0);
            
            expect(await brlt.allowance(owner.address, user1.address)).to.equal(0);
        });

        it("Should allow permit with zero value (sets allowance to 0, emits event)", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(owner.address, mintAmount); // owner has tokens

            // First, set some allowance via permit
            const deadline = Math.floor(Date.now() / 1000) + 3600;
            let nonce = await brlt.nonces(owner.address);
            const {v: v1, r: r1, s: s1} = ethers.Signature.from( (await getPermitSignature(owner, brlt, user1.address, H(100), deadline, nonce)).signature );
            await brlt.permit(owner.address, user1.address, H(100), deadline, v1, r1, s1);
            expect(await brlt.allowance(owner.address, user1.address)).to.equal(H(100));
            nonce = await brlt.nonces(owner.address); // nonce has incremented

            // Then, permit zero value
            const {v: v2, r: r2, s: s2} = ethers.Signature.from( (await getPermitSignature(owner, brlt, user1.address, 0, deadline, nonce)).signature );
            await expect(brlt.permit(owner.address, user1.address, 0, deadline, v2, r2, s2))
                .to.emit(brlt, "Approval")
                .withArgs(owner.address, user1.address, 0);
            
            expect(await brlt.allowance(owner.address, user1.address)).to.equal(0);
            expect(await brlt.nonces(owner.address)).to.equal(nonce + 1n);
        });

        it("Should allow transferring zero tokens (no state change, emits event)", async function () {
            const { brlt, owner, user1, user2 } = await loadFixture(deployBRLTFixture);
            await brlt.connect(owner).mint(user1.address, mintAmount);

            const user1InitialBalance = await brlt.balanceOf(user1.address);
            const user2InitialBalance = await brlt.balanceOf(user2.address);

            await expect(brlt.connect(user1).transfer(user2.address, 0))
                .to.emit(brlt, "Transfer")
                .withArgs(user1.address, user2.address, 0);
            
            expect(await brlt.balanceOf(user1.address)).to.equal(user1InitialBalance);
            expect(await brlt.balanceOf(user2.address)).to.equal(user2InitialBalance);
        });

    });

    describe("UUPS Upgradeability", function () {
        const mintAmount = H(100);

        it("Should allow UPGRADER_ROLE to upgrade the contract", async function () {
            const { brlt, owner, UPGRADER_ROLE } = await loadFixture(deployBRLTFixture);
            expect(await brlt.hasRole(UPGRADER_ROLE, owner.address)).to.be.true;

            const BRLTv2Factory = await ethers.getContractFactory("BRLTv2");
            // const brltV2Impl = await BRLTv2Factory.deploy(); // Not needed to deploy impl separately for upgradeProxy
            // await brltV2Impl.waitForDeployment();

            const brltProxyAddress = await brlt.getAddress();

            const brltUpgraded = await upgrades.upgradeProxy(brltProxyAddress, BRLTv2Factory, { kind: 'uups' });
            await brltUpgraded.waitForDeployment();

            expect(await brltUpgraded.symbol()).to.equal("BRLTV2");

            const v2Value = 999;
            // Need to connect to the signer who has the role to call initializeV2, if it's permissioned
            // BRLTv2's initializeV2 is public, so owner can call.
            await expect(brltUpgraded.connect(owner).initializeV2(v2Value))
                .to.not.be.reverted;
            expect(await brltUpgraded.getVersion2Field()).to.equal(v2Value);
        });

        it("Should NOT allow an account without UPGRADER_ROLE to upgrade", async function () {
            const { brlt, user1, UPGRADER_ROLE } = await loadFixture(deployBRLTFixture);
            const BRLTv2Factory = await ethers.getContractFactory("BRLTv2");
            const brltProxyAddress = await brlt.getAddress();

            expect(await brlt.hasRole(UPGRADER_ROLE, user1.address)).to.be.false;

            await expect(upgrades.upgradeProxy(brltProxyAddress, BRLTv2Factory.connect(user1), { kind: 'uups' }))
                .to.be.revertedWithCustomError(brlt, "AccessControlUnauthorizedAccount")
                .withArgs(user1.address, UPGRADER_ROLE);
        });

        it("Should preserve state (e.g., balances, roles) after upgrade", async function () {
            const { brlt, owner, user1, MINTER_ROLE } = await loadFixture(deployBRLTFixture);
            
            await brlt.connect(owner).mint(user1.address, mintAmount);
            expect(await brlt.balanceOf(user1.address)).to.equal(mintAmount);
            expect(await brlt.hasRole(MINTER_ROLE, owner.address)).to.be.true;

            const BRLTv2Factory = await ethers.getContractFactory("BRLTv2");
            const brltProxyAddress = await brlt.getAddress();
            const brltUpgraded = await upgrades.upgradeProxy(brltProxyAddress, BRLTv2Factory, { kind: 'uups' });
            await brltUpgraded.waitForDeployment();

            expect(await brltUpgraded.balanceOf(user1.address)).to.equal(mintAmount);
            expect(await brltUpgraded.hasRole(MINTER_ROLE, owner.address)).to.be.true;
            expect(await brltUpgraded.name()).to.equal("BRLT");
            expect(await brltUpgraded.symbol()).to.equal("BRLTV2");
        });

        it("V1 functionality should still work on upgraded contract (if not overridden)", async function () {
            const { brlt, owner, user1 } = await loadFixture(deployBRLTFixture);
            const BRLTv2Factory = await ethers.getContractFactory("BRLTv2");
            const brltProxyAddress = await brlt.getAddress();
            const brltUpgraded = await upgrades.upgradeProxy(brltProxyAddress, BRLTv2Factory, { kind: 'uups' });
            await brltUpgraded.waitForDeployment();

            await expect(brltUpgraded.connect(owner).mint(user1.address, mintAmount))
                .to.emit(brltUpgraded, "Transfer")
                .withArgs(ethers.ZeroAddress, user1.address, mintAmount);
            expect(await brltUpgraded.balanceOf(user1.address)).to.equal(mintAmount);
        });

        it("Should prevent upgrade if new implementation is not UUPS compatible (e.g., missing _authorizeUpgrade or proxiableUUID)", async function() {
            const { brlt, owner } = await loadFixture(deployBRLTFixture);
            const NonUpgradeableFactory = await ethers.getContractFactory("MockToken");
            const brltProxyAddress = await brlt.getAddress();

            try {
                await upgrades.upgradeProxy(brltProxyAddress, NonUpgradeableFactory.connect(owner), { kind: 'uups' });
                expect.fail("Upgrade to non-safe contract should have failed");
            } catch (error) {
                expect(error.message).to.include("is not upgrade safe");
            }
        });
    });

    // Add more describe blocks for other specific scenarios or edge cases if needed
}); 