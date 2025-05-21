const { ethers } = require("hardhat");
const { expect } = require("chai");

const { loadFixture } = require("@nomicfoundation/hardhat-network-helpers");
const { upgrades } = require("hardhat");

// Helper function to convert amounts to the token's decimal representation
function H(amount) {
    return ethers.parseUnits(amount.toString(), 6); // BRLT now has 6 decimals
}

describe("BRLT Integration Scenarios", function () {
    // We define a fixture to reuse the same setup in every test.
    // We use loadFixture to run this setup once, snapshot that state,
    // and reset Hardhat Network to that snapshot in every test.
    async function deployBRLTFixture() {
        // Contracts are deployed using the first signer/account by default
        const [owner, admin, minter, burner, pauser, blacklister, upgrader, userA, userB, userC, userD, spenderC]
            = await ethers.getSigners();

        const BRLT = await ethers.getContractFactory("BRLT", owner);
        // Deploy as an upgradeable proxy
        const brlt = await upgrades.deployProxy(BRLT, [admin.address], { 
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
        
        // Initial role setup: admin has all roles from deployment
        // Grant specific roles to dedicated accounts for clarity in tests
        await brlt.connect(admin).grantRole(MINTER_ROLE, minter.address);
        await brlt.connect(admin).grantRole(BURNER_ROLE, burner.address);
        await brlt.connect(admin).grantRole(PAUSER_ROLE, pauser.address);
        await brlt.connect(admin).grantRole(BLACKLISTER_ROLE, blacklister.address);
        await brlt.connect(admin).grantRole(UPGRADER_ROLE, upgrader.address);

        return { 
            brlt, 
            owner, admin, minter, burner, pauser, blacklister, upgrader, 
            userA, userB, userC, userD, spenderC,
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

    describe("Full User Lifecycle Scenario", function () {
        it("should allow a complete user journey: grant roles, mint, transfer, approve, transferFrom, permit, burn", async function () {
            const { brlt, admin, minter, burner, userA, userB, spenderC, userD, MINTER_ROLE, BURNER_ROLE } = await loadFixture(deployBRLTFixture);
            
            const initialMintAmount = H(1000);
            const transferAmount1 = H(300);
            const approvalAmount = H(100);
            const transferFromAmount = H(50);
            const permitAmount = H(20);
            const burnAmount = H(10);

            // 1. Admin already granted MINTER_ROLE to minter in fixture.
            // Verify minter has MINTER_ROLE (optional check, good for sanity)
            expect(await brlt.hasRole(MINTER_ROLE, minter.address)).to.be.true;

            // 2. Minter mints tokens to UserA.
            await expect(brlt.connect(minter).mint(userA.address, initialMintAmount))
                .to.emit(brlt, "Transfer").withArgs(ethers.ZeroAddress, userA.address, initialMintAmount);
            expect(await brlt.balanceOf(userA.address)).to.equal(initialMintAmount);
            console.log(`UserA balance after mint: ${await brlt.balanceOf(userA.address)}`);

            // 3. UserA transfers some tokens to UserB.
            await expect(brlt.connect(userA).transfer(userB.address, transferAmount1))
                .to.emit(brlt, "Transfer").withArgs(userA.address, userB.address, transferAmount1);
            expect(await brlt.balanceOf(userA.address)).to.equal(initialMintAmount - transferAmount1);
            expect(await brlt.balanceOf(userB.address)).to.equal(transferAmount1);
            console.log(`UserA balance after transfer to UserB: ${await brlt.balanceOf(userA.address)}`);
            console.log(`UserB balance after transfer from UserA: ${await brlt.balanceOf(userB.address)}`);

            // 4. UserA approves SpenderC to spend some tokens.
            await expect(brlt.connect(userA).approve(spenderC.address, approvalAmount))
                .to.emit(brlt, "Approval").withArgs(userA.address, spenderC.address, approvalAmount);
            expect(await brlt.allowance(userA.address, spenderC.address)).to.equal(approvalAmount);
            console.log(`SpenderC allowance from UserA: ${await brlt.allowance(userA.address, spenderC.address)}`);

            // 5. SpenderC uses transferFrom to move tokens from UserA to UserD.
            await expect(brlt.connect(spenderC).transferFrom(userA.address, userD.address, transferFromAmount))
                .to.emit(brlt, "Transfer").withArgs(userA.address, userD.address, transferFromAmount);
            expect(await brlt.balanceOf(userA.address)).to.equal(initialMintAmount - transferAmount1 - transferFromAmount);
            expect(await brlt.balanceOf(userD.address)).to.equal(transferFromAmount);
            expect(await brlt.allowance(userA.address, spenderC.address)).to.equal(approvalAmount - transferFromAmount);
            console.log(`UserA balance after transferFrom by SpenderC: ${await brlt.balanceOf(userA.address)}`);
            console.log(`UserD balance after transferFrom by SpenderC: ${await brlt.balanceOf(userD.address)}`);

            // 6. UserA uses permit to approve SpenderC for another amount.
            const deadline = Math.floor(Date.now() / 1000) + 3600;
            const nonce = await brlt.nonces(userA.address);
            const permitSig = await getPermitSignature(userA, brlt, spenderC.address, permitAmount, deadline, nonce);
            
            // const { r, s, v } = ethers.Signature.from(permitSig.signature); // Correct way to split signature
            // Using ethers.splitSignature as ethers.Signature.from() might not be available in all versions directly for splitting.
            const sig = ethers.Signature.from(permitSig.signature); // Or ethers.splitSignature(permitSig.signature)

            await expect(brlt.connect(userA).permit(userA.address, spenderC.address, permitAmount, deadline, sig.v, sig.r, sig.s))
                 .to.emit(brlt, "Approval").withArgs(userA.address, spenderC.address, permitAmount);
            expect(await brlt.allowance(userA.address, spenderC.address)).to.equal(permitAmount);
            console.log(`SpenderC allowance from UserA after permit: ${await brlt.allowance(userA.address, spenderC.address)}`);

            // 7. Admin already granted BURNER_ROLE to burner in fixture.
            // Verify burner has BURNER_ROLE (optional check)
            expect(await brlt.hasRole(BURNER_ROLE, burner.address)).to.be.true;

            // 8. UserA allows the burner account to burn some of their tokens (approve then burner calls burnFrom).
            // First, UserA approves the burner.
            await expect(brlt.connect(userA).approve(burner.address, burnAmount))
                .to.emit(brlt, "Approval").withArgs(userA.address, burner.address, burnAmount);
            expect(await brlt.allowance(userA.address, burner.address)).to.equal(burnAmount);
            
            // Then, burner burns from UserA.
            const userABalanceBeforeBurn = await brlt.balanceOf(userA.address);
            await expect(brlt.connect(burner).burnFrom(userA.address, burnAmount))
                .to.emit(brlt, "Transfer").withArgs(userA.address, ethers.ZeroAddress, burnAmount);
            expect(await brlt.balanceOf(userA.address)).to.equal(userABalanceBeforeBurn - burnAmount);
            console.log(`UserA balance after burn: ${await brlt.balanceOf(userA.address)}`);
        });
    });

    describe("Pause and Resume Scenario", function () {
        it("should restrict and allow operations when paused and unpaused", async function () {
            const { brlt, admin, minter, burner, pauser, userA, userB, PAUSER_ROLE } = await loadFixture(deployBRLTFixture);

            const mintAmount = H(1000);
            const transferAmount = H(100);
            const burnAmount = H(50);

            // Setup: Mint some tokens to UserA
            await brlt.connect(minter).mint(userA.address, mintAmount);
            expect(await brlt.balanceOf(userA.address)).to.equal(mintAmount);

            // 1. Pauser (who has PAUSER_ROLE from fixture) pauses the contract
            expect(await brlt.hasRole(PAUSER_ROLE, pauser.address)).to.be.true;
            await expect(brlt.connect(pauser).pause())
                .to.emit(brlt, "Paused").withArgs(pauser.address);
            expect(await brlt.paused()).to.be.true;
            console.log("Contract paused.");

            // 2. Verify operations fail while paused
            // Transfer
            await expect(brlt.connect(userA).transfer(userB.address, transferAmount))
                .to.be.revertedWithCustomError(brlt, "EnforcedPause");
            // Mint
            await expect(brlt.connect(minter).mint(userA.address, mintAmount))
                .to.be.revertedWithCustomError(brlt, "EnforcedPause");
            // BurnFrom (UserA approves burner, burner tries to burn)
            await brlt.connect(userA).approve(burner.address, burnAmount); // Approve should still work
            await expect(brlt.connect(burner).burnFrom(userA.address, burnAmount))
                .to.be.revertedWithCustomError(brlt, "EnforcedPause");
            console.log("Operations correctly restricted while paused.");

            // 3. Pauser unpauses the contract
            await expect(brlt.connect(pauser).unpause())
                .to.emit(brlt, "Unpaused").withArgs(pauser.address);
            expect(await brlt.paused()).to.be.false;
            console.log("Contract unpaused.");

            // 4. Verify operations succeed after unpausing
            // Transfer
            const userABalanceBeforeTransfer = await brlt.balanceOf(userA.address);
            await expect(brlt.connect(userA).transfer(userB.address, transferAmount))
                .to.emit(brlt, "Transfer").withArgs(userA.address, userB.address, transferAmount);
            expect(await brlt.balanceOf(userA.address)).to.equal(userABalanceBeforeTransfer - transferAmount);
            expect(await brlt.balanceOf(userB.address)).to.equal(transferAmount); // Assuming userB had 0 before

            // Mint
            const totalSupplyBeforeMint = await brlt.totalSupply();
            const userABalanceBeforeMint = await brlt.balanceOf(userA.address);
            await expect(brlt.connect(minter).mint(userA.address, mintAmount))
                .to.emit(brlt, "Transfer").withArgs(ethers.ZeroAddress, userA.address, mintAmount);
            expect(await brlt.totalSupply()).to.equal(totalSupplyBeforeMint + mintAmount);
            expect(await brlt.balanceOf(userA.address)).to.equal(userABalanceBeforeMint + mintAmount);

            // BurnFrom
            // UserA already approved burner. Let's ensure allowance is sufficient for a new burnAmount.
            await brlt.connect(userA).approve(burner.address, burnAmount); 
            const userABalanceBeforeBurn = await brlt.balanceOf(userA.address);
            const totalSupplyBeforeBurn = await brlt.totalSupply();
            await expect(brlt.connect(burner).burnFrom(userA.address, burnAmount))
                .to.emit(brlt, "Transfer").withArgs(userA.address, ethers.ZeroAddress, burnAmount);
            expect(await brlt.balanceOf(userA.address)).to.equal(userABalanceBeforeBurn - burnAmount);
            expect(await brlt.totalSupply()).to.equal(totalSupplyBeforeBurn - burnAmount);
            console.log("Operations correctly allowed after unpausing.");
        });
    });

    describe("Blacklisting Scenario", function () {
        it("should restrict and allow operations for blacklisted accounts", async function () {
            const { brlt, admin, minter, blacklister, userA, userB, userC, BLACKLISTER_ROLE } = await loadFixture(deployBRLTFixture);

            const mintAmount = H(1000);
            const transferAmount = H(100);

            // Setup: Mint tokens to UserA and UserB
            await brlt.connect(minter).mint(userA.address, mintAmount);
            await brlt.connect(minter).mint(userB.address, mintAmount);
            expect(await brlt.balanceOf(userA.address)).to.equal(mintAmount);
            expect(await brlt.balanceOf(userB.address)).to.equal(mintAmount);

            // 1. Blacklister (who has BLACKLISTER_ROLE from fixture) blacklists UserA
            expect(await brlt.hasRole(BLACKLISTER_ROLE, blacklister.address)).to.be.true;
            await expect(brlt.connect(blacklister).blacklistAddress(userA.address))
                .to.emit(brlt, "AddressBlacklistedStatusChanged").withArgs(userA.address, true);
            expect(await brlt.isBlacklisted(userA.address)).to.be.true;
            console.log(`UserA blacklisted: ${await brlt.isBlacklisted(userA.address)}`);

            // 2. Verify UserA (blacklisted) cannot send tokens
            await expect(brlt.connect(userA).transfer(userB.address, transferAmount))
                .to.be.revertedWithCustomError(brlt, "AccountBlacklisted").withArgs(userA.address);

            // 3. Verify UserA (blacklisted) cannot receive tokens (UserB tries to send to UserA)
            await expect(brlt.connect(userB).transfer(userA.address, transferAmount))
                .to.be.revertedWithCustomError(brlt, "AccountBlacklisted").withArgs(userA.address);
            
            // 4. Verify UserB (not blacklisted) can still transact with another non-blacklisted user (UserC)
            // Mint to UserC first for this check, or ensure UserC gets tokens another way if preferred.
            // For simplicity, let's assume UserC doesn't need prior balance for this specific check of UserB's ability to send.
            // If UserC had tokens, we could also check UserB receiving from UserC.
            await expect(brlt.connect(userB).transfer(userC.address, transferAmount))
                .to.emit(brlt, "Transfer").withArgs(userB.address, userC.address, transferAmount);
            expect(await brlt.balanceOf(userB.address)).to.equal(mintAmount - transferAmount);
            expect(await brlt.balanceOf(userC.address)).to.equal(transferAmount);
            console.log("UserB (not blacklisted) successfully transferred to UserC.");

            // 5. Blacklister unblacklists UserA
            await expect(brlt.connect(blacklister).unblacklistAddress(userA.address))
                .to.emit(brlt, "AddressBlacklistedStatusChanged").withArgs(userA.address, false);
            expect(await brlt.isBlacklisted(userA.address)).to.be.false;
            console.log(`UserA blacklisted: ${await brlt.isBlacklisted(userA.address)}`);

            // 6. Verify UserA can send tokens again
            const userABalanceBeforeSend = await brlt.balanceOf(userA.address);
            await expect(brlt.connect(userA).transfer(userB.address, transferAmount))
                .to.emit(brlt, "Transfer").withArgs(userA.address, userB.address, transferAmount);
            expect(await brlt.balanceOf(userA.address)).to.equal(userABalanceBeforeSend - transferAmount);
            // UserB's balance will increase by transferAmount from its state before this transfer
            const userBBalanceAfterUnblacklistTransfer = await brlt.balanceOf(userB.address);
            expect(userBBalanceAfterUnblacklistTransfer).to.equal((mintAmount - transferAmount) + transferAmount); 

            // 7. Verify UserA can receive tokens again (UserB sends back to UserA)
            const userBBalanceBeforeSendToA = await brlt.balanceOf(userB.address);
            const userABalanceBeforeReceive = await brlt.balanceOf(userA.address);
            await expect(brlt.connect(userB).transfer(userA.address, transferAmount))
                .to.emit(brlt, "Transfer").withArgs(userB.address, userA.address, transferAmount);
            expect(await brlt.balanceOf(userB.address)).to.equal(userBBalanceBeforeSendToA - transferAmount);
            expect(await brlt.balanceOf(userA.address)).to.equal(userABalanceBeforeReceive + transferAmount);
            console.log("UserA can send and receive tokens again after unblacklisting.");
        });
    });

    describe("Access Control Management Scenario", function () {
        it("should correctly manage role granting, revoking, and renouncing", async function () {
            const { 
                brlt, admin, minter, pauser, userA, userB,
                MINTER_ROLE, PAUSER_ROLE, DEFAULT_ADMIN_ROLE 
            } = await loadFixture(deployBRLTFixture);

            const mintAmount = H(100);

            // Initial state: userA has no MINTER_ROLE or PAUSER_ROLE
            expect(await brlt.hasRole(MINTER_ROLE, userA.address)).to.be.false;
            expect(await brlt.hasRole(PAUSER_ROLE, userA.address)).to.be.false;

            // 1. Admin (has DEFAULT_ADMIN_ROLE from fixture) grants MINTER_ROLE to userA
            // Note: admin already granted MINTER_ROLE to 'minter' and PAUSER_ROLE to 'pauser' in fixture.
            // We'll use userA for a fresh grant/revoke cycle.
            await expect(brlt.connect(admin).grantRole(MINTER_ROLE, userA.address))
                .to.emit(brlt, "RoleGranted").withArgs(MINTER_ROLE, userA.address, admin.address);
            expect(await brlt.hasRole(MINTER_ROLE, userA.address)).to.be.true;
            console.log("Admin granted MINTER_ROLE to UserA.");

            // 2. UserA (now a minter) can mint tokens
            await expect(brlt.connect(userA).mint(userB.address, mintAmount))
                .to.emit(brlt, "Transfer").withArgs(ethers.ZeroAddress, userB.address, mintAmount);
            expect(await brlt.balanceOf(userB.address)).to.equal(mintAmount);
            console.log("UserA successfully minted tokens as MINTER.");

            // 3. Admin revokes MINTER_ROLE from userA
            await expect(brlt.connect(admin).revokeRole(MINTER_ROLE, userA.address))
                .to.emit(brlt, "RoleRevoked").withArgs(MINTER_ROLE, userA.address, admin.address);
            expect(await brlt.hasRole(MINTER_ROLE, userA.address)).to.be.false;
            console.log("Admin revoked MINTER_ROLE from UserA.");

            // 4. UserA can no longer mint tokens
            await expect(brlt.connect(userA).mint(userB.address, mintAmount))
                .to.be.revertedWithCustomError(brlt, "AccessControlUnauthorizedAccount")
                .withArgs(userA.address, MINTER_ROLE);
            console.log("UserA failed to mint tokens after role revocation (expected).");

            // 5. Grant PAUSER_ROLE to userA by admin
            await expect(brlt.connect(admin).grantRole(PAUSER_ROLE, userA.address))
                .to.emit(brlt, "RoleGranted").withArgs(PAUSER_ROLE, userA.address, admin.address);
            expect(await brlt.hasRole(PAUSER_ROLE, userA.address)).to.be.true;
            console.log("Admin granted PAUSER_ROLE to UserA.");

            // 6. UserA (now a pauser) can pause the contract
            await expect(brlt.connect(userA).pause())
                .to.emit(brlt, "Paused").withArgs(userA.address);
            expect(await brlt.paused()).to.be.true;
            console.log("UserA successfully paused the contract as PAUSER.");

            // 7. UserA renounces their PAUSER_ROLE
            // UserA must be connected to renounce their own role
            await expect(brlt.connect(userA).renounceRole(PAUSER_ROLE, userA.address))
                .to.emit(brlt, "RoleRevoked").withArgs(PAUSER_ROLE, userA.address, userA.address);
            expect(await brlt.hasRole(PAUSER_ROLE, userA.address)).to.be.false;
            console.log("UserA renounced PAUSER_ROLE.");

            // 8. UserA can no longer unpause (or pause) the contract
            // Need to use a different pauser (e.g. the original 'pauser' account from fixture) or admin to unpause first if we want to test pause again.
            // For this test, admin (who has PAUSER_ROLE) will unpause.
             await expect(brlt.connect(admin).unpause())
                .to.emit(brlt, "Unpaused").withArgs(admin.address);
            expect(await brlt.paused()).to.be.false;
            
            await expect(brlt.connect(userA).pause()) // UserA tries to pause again
                .to.be.revertedWithCustomError(brlt, "AccessControlUnauthorizedAccount")
                .withArgs(userA.address, PAUSER_ROLE);
            console.log("UserA failed to pause contract after renouncing role (expected).");
            
            // 9. UserB (no DEFAULT_ADMIN_ROLE) tries to grant MINTER_ROLE to userA
            expect(await brlt.hasRole(DEFAULT_ADMIN_ROLE, userB.address)).to.be.false;
            await expect(brlt.connect(userB).grantRole(MINTER_ROLE, userA.address))
                .to.be.revertedWithCustomError(brlt, "AccessControlUnauthorizedAccount")
                .withArgs(userB.address, DEFAULT_ADMIN_ROLE);
            console.log("UserB failed to grant MINTER_ROLE (not admin, expected).");

            // 10. Check if default 'minter' account (from fixture) still has MINTER_ROLE (it should)
            expect(await brlt.hasRole(MINTER_ROLE, minter.address)).to.be.true;
            await expect(brlt.connect(minter).mint(userB.address, mintAmount))
                .to.emit(brlt, "Transfer");
            console.log("Original 'minter' account can still mint.");
        });
    });

    describe("Complex Upgrade Scenario (V1 -> V2 -> V3)", function () {
        it("should preserve state and functionality across multiple upgrades", async function () {
            const { 
                brlt, owner, admin, minter, burner, pauser, blacklister, upgrader, 
                userA, userB, userC, spenderC,
                MINTER_ROLE, PAUSER_ROLE, BLACKLISTER_ROLE, UPGRADER_ROLE
            } = await loadFixture(deployBRLTFixture); // BRLT (V1) is deployed here

            const initialMintAmount = H(1000);
            const v1ApprovalAmount = H(500);
            const v2FieldToSet = 999;
            const v3ReportBlock = 12345;

            // --- Initial V1 Operations & State Setup ---
            console.log("V1: Performing initial operations...");
            // Roles already set up in fixture: minter, pauser, blacklister, upgrader have their roles from admin.
            // admin is also owner and has all roles initially.

            await brlt.connect(minter).mint(userA.address, initialMintAmount);
            await brlt.connect(pauser).pause();
            await brlt.connect(blacklister).blacklistAddress(userB.address);
            await brlt.connect(userA).approve(spenderC.address, v1ApprovalAmount);

            const v1UserABalance = await brlt.balanceOf(userA.address);
            const v1UserBBalance = await brlt.balanceOf(userB.address); // Should be 0
            const v1UserCBalance = await brlt.balanceOf(userC.address); // Should be 0
            const v1AllowanceUserASpenderC = await brlt.allowance(userA.address, spenderC.address);
            const v1PausedState = await brlt.paused();
            const v1IsUserBBlacklisted = await brlt.isBlacklisted(userB.address);
            const v1AdminHasMinterRole = await brlt.hasRole(MINTER_ROLE, admin.address); // From fixture
            const v1MinterHasMinterRole = await brlt.hasRole(MINTER_ROLE, minter.address); // From fixture

            expect(v1UserABalance).to.equal(initialMintAmount);
            expect(v1AllowanceUserASpenderC).to.equal(v1ApprovalAmount);
            expect(v1PausedState).to.be.true;
            expect(v1IsUserBBlacklisted).to.be.true;

            // --- Upgrade to V2 ---
            console.log("Upgrading V1 to V2...");
            const BRLTv2 = await ethers.getContractFactory("BRLTv2", upgrader); // Connect upgrader
            const brltProxyAddress = await brlt.getAddress();
            const brltV2 = await upgrades.upgradeProxy(brltProxyAddress, BRLTv2, { kind: 'uups' });
            await brltV2.waitForDeployment();
            console.log("Upgraded to V2 at address:", await brltV2.getAddress());

            // --- Verify V1 State Preservation in V2 & V2 Functionality ---
            console.log("V2: Verifying V1 state and V2 functionality...");
            expect(await brltV2.balanceOf(userA.address)).to.equal(v1UserABalance);
            expect(await brltV2.balanceOf(userB.address)).to.equal(v1UserBBalance);
            expect(await brltV2.allowance(userA.address, spenderC.address)).to.equal(v1AllowanceUserASpenderC);
            expect(await brltV2.paused()).to.equal(v1PausedState); // Still paused
            expect(await brltV2.isBlacklisted(userB.address)).to.equal(v1IsUserBBlacklisted);
            expect(await brltV2.hasRole(MINTER_ROLE, admin.address)).to.equal(v1AdminHasMinterRole);
            expect(await brltV2.hasRole(MINTER_ROLE, minter.address)).to.equal(v1MinterHasMinterRole);
            expect(await brltV2.name()).to.equal("BRLT"); // Name from V1
            expect(await brltV2.symbol()).to.equal("BRLTV2"); // Symbol from V2

            // Initialize V2
            await expect(brltV2.connect(upgrader).initializeV2(v2FieldToSet))
                .to.emit(brltV2, "V2Initialized").withArgs(upgrader.address, v2FieldToSet);
            expect(await brltV2.getVersion2Field()).to.equal(v2FieldToSet);

            // Unpause to test V1 transfer functionality in V2
            await brltV2.connect(pauser).unpause(); // Pauser from fixture
            expect(await brltV2.paused()).to.be.false;

            const v2TransferAmount = H(100);
            await expect(brltV2.connect(userA).transfer(userC.address, v2TransferAmount))
                .to.emit(brltV2, "Transfer").withArgs(userA.address, userC.address, v2TransferAmount);
            expect(await brltV2.balanceOf(userA.address)).to.equal(v1UserABalance - v2TransferAmount);
            expect(await brltV2.balanceOf(userC.address)).to.equal(v1UserCBalance + v2TransferAmount);

            // --- Upgrade to V3 ---
            console.log("Upgrading V2 to V3...");
            const BRLTv3 = await ethers.getContractFactory("BRLTv3", upgrader); // Connect upgrader
            const brltV3 = await upgrades.upgradeProxy(await brltV2.getAddress(), BRLTv3, { kind: 'uups' });
            await brltV3.waitForDeployment();
            console.log("Upgraded to V3 at address:", await brltV3.getAddress());

            // --- Verify V1 & V2 State Preservation in V3 & V3 Functionality ---
            console.log("V3: Verifying V1/V2 state and V3 functionality...");
            expect(await brltV3.balanceOf(userA.address)).to.equal(v1UserABalance - v2TransferAmount);
            expect(await brltV3.balanceOf(userC.address)).to.equal(v1UserCBalance + v2TransferAmount);
            expect(await brltV3.allowance(userA.address, spenderC.address)).to.equal(v1AllowanceUserASpenderC);
            expect(await brltV3.paused()).to.be.false; // Was unpaused in V2 state
            expect(await brltV3.isBlacklisted(userB.address)).to.equal(v1IsUserBBlacklisted);
            expect(await brltV3.hasRole(MINTER_ROLE, minter.address)).to.equal(v1MinterHasMinterRole);
            expect(await brltV3.getVersion2Field()).to.equal(v2FieldToSet); // From V2
            expect(await brltV3.name()).to.equal("BRLT"); // Name from V1
            expect(await brltV3.symbol()).to.equal("BRLTV3"); // Symbol from V3

            // Initialize V3 - upgrader calls it and gets REPORTER_ROLE
            await expect(brltV3.connect(upgrader).initializeV3())
                .to.emit(brltV3, "V3Initialized").withArgs(upgrader.address);
            const REPORTER_ROLE = await brltV3.REPORTER_ROLE();
            expect(await brltV3.hasRole(REPORTER_ROLE, upgrader.address)).to.be.true;

            // Test V3 specific function: makeReport
            await expect(brltV3.connect(upgrader).makeReport(v3ReportBlock))
                .to.emit(brltV3, "ReportMade").withArgs(upgrader.address, v3ReportBlock);
            expect(await brltV3.lastReportedBlock()).to.equal(v3ReportBlock);

            // Test V3 view function combining V2 and V3 state
            const reportData = await brltV3.getReportData();            
            expect(reportData.v2Field).to.equal(v2FieldToSet);
            expect(reportData.reportedBlock).to.equal(v3ReportBlock);

            // Test access control for makeReport
            await expect(brltV3.connect(userA).makeReport(v3ReportBlock + 1))
                .to.be.revertedWithCustomError(brltV3, "AccessControlUnauthorizedAccount")
                .withArgs(userA.address, REPORTER_ROLE);
            
            console.log("Complex upgrade scenario completed successfully.");
        });
    });

    // TODO: Add describe blocks for other scenarios:
    // - Upgrade Scenario (more complex than unit test)
}); 