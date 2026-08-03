// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

/// @title LCVEscrow
/// @notice Milestone escrow for real-estate / off-plan sales. A buyer deposits the full sale
/// price up front; a neutral arbiter (typically an escrow agent or the developer's regulator)
/// releases funds to the seller milestone by milestone as construction/delivery progresses,
/// or refunds whatever remains undisbursed to the buyer if the deal falls through.
/// @dev Reference contract for LastcoinVision's EVM compatibility layer (Phase D). Native-token
/// (LCV) denominated — no ERC-20 dependency, since LCV is directly usable as msg.value via the
/// chain's own precisebank bridge.
contract LCVEscrow {
    struct Milestone {
        string description;
        uint256 amount;
        bool released;
    }

    struct Escrow {
        address buyer;
        address seller;
        address arbiter;
        uint256 totalAmount;
        uint256 deposited;
        uint256 releasedAmount;
        bool cancelled;
        Milestone[] milestones;
    }

    uint256 public nextEscrowId;
    mapping(uint256 => Escrow) private escrows;

    event EscrowCreated(uint256 indexed escrowId, address indexed buyer, address indexed seller, address arbiter, uint256 totalAmount);
    event Deposited(uint256 indexed escrowId, address indexed buyer, uint256 amount);
    event MilestoneReleased(uint256 indexed escrowId, uint256 indexed milestoneIndex, uint256 amount);
    event EscrowCancelled(uint256 indexed escrowId, uint256 refundAmount);

    error NotArbiter();
    error NotBuyer();
    error EscrowIsCancelled();
    error NoMilestones();
    error AmountMismatch();
    error AlreadyFullyDeposited();
    error NotFullyDeposited();
    error MilestoneAlreadyReleased();
    error InvalidMilestoneIndex();
    error NothingToRefund();
    error ZeroAddress();

    modifier onlyArbiter(uint256 escrowId) {
        if (msg.sender != escrows[escrowId].arbiter) revert NotArbiter();
        _;
    }

    /// @notice Creates a new milestone escrow. The total amount is the sum of all milestone
    /// amounts; the buyer deposits it via `deposit` in one or more installments.
    function createEscrow(
        address seller,
        address arbiter,
        string[] calldata descriptions,
        uint256[] calldata amounts
    ) external returns (uint256 escrowId) {
        if (seller == address(0) || arbiter == address(0)) revert ZeroAddress();
        if (descriptions.length == 0 || descriptions.length != amounts.length) revert NoMilestones();

        escrowId = nextEscrowId++;
        Escrow storage e = escrows[escrowId];
        e.buyer = msg.sender;
        e.seller = seller;
        e.arbiter = arbiter;

        uint256 total = 0;
        for (uint256 i = 0; i < amounts.length; i++) {
            if (amounts[i] == 0) revert AmountMismatch();
            total += amounts[i];
            e.milestones.push(Milestone({description: descriptions[i], amount: amounts[i], released: false}));
        }
        e.totalAmount = total;

        emit EscrowCreated(escrowId, msg.sender, seller, arbiter, total);
    }

    /// @notice The buyer deposits funds toward the escrow's total amount. Deposits beyond the
    /// total are rejected outright so the contract never holds more than the agreed price.
    function deposit(uint256 escrowId) external payable {
        Escrow storage e = escrows[escrowId];
        if (e.cancelled) revert EscrowIsCancelled();
        if (msg.sender != e.buyer) revert NotBuyer();
        if (e.deposited + msg.value > e.totalAmount) revert AmountMismatch();
        if (e.deposited == e.totalAmount) revert AlreadyFullyDeposited();

        e.deposited += msg.value;
        emit Deposited(escrowId, msg.sender, msg.value);
    }

    /// @notice The arbiter releases a single milestone's funds to the seller once its
    /// deliverable has been confirmed. Requires the escrow to be fully funded first, so a
    /// milestone is never paid out of an under-collateralized deal.
    function releaseMilestone(uint256 escrowId, uint256 milestoneIndex) external onlyArbiter(escrowId) {
        Escrow storage e = escrows[escrowId];
        if (e.cancelled) revert EscrowIsCancelled();
        if (e.deposited != e.totalAmount) revert NotFullyDeposited();
        if (milestoneIndex >= e.milestones.length) revert InvalidMilestoneIndex();

        Milestone storage m = e.milestones[milestoneIndex];
        if (m.released) revert MilestoneAlreadyReleased();

        m.released = true;
        e.releasedAmount += m.amount;

        emit MilestoneReleased(escrowId, milestoneIndex, m.amount);
        (bool ok, ) = e.seller.call{value: m.amount}("");
        require(ok, "transfer to seller failed");
    }

    /// @notice The arbiter cancels the escrow and refunds whatever has been deposited but not
    /// yet released back to the buyer — the deal-falls-through path.
    function cancelAndRefund(uint256 escrowId) external onlyArbiter(escrowId) {
        Escrow storage e = escrows[escrowId];
        if (e.cancelled) revert EscrowIsCancelled();

        uint256 refundAmount = e.deposited - e.releasedAmount;
        e.cancelled = true;

        emit EscrowCancelled(escrowId, refundAmount);
        if (refundAmount == 0) revert NothingToRefund();

        (bool ok, ) = e.buyer.call{value: refundAmount}("");
        require(ok, "refund to buyer failed");
    }

    function getEscrow(uint256 escrowId)
        external
        view
        returns (address buyer, address seller, address arbiter, uint256 totalAmount, uint256 deposited, uint256 releasedAmount, bool cancelled, uint256 milestoneCount)
    {
        Escrow storage e = escrows[escrowId];
        return (e.buyer, e.seller, e.arbiter, e.totalAmount, e.deposited, e.releasedAmount, e.cancelled, e.milestones.length);
    }

    function getMilestone(uint256 escrowId, uint256 milestoneIndex)
        external
        view
        returns (string memory description, uint256 amount, bool released)
    {
        Milestone storage m = escrows[escrowId].milestones[milestoneIndex];
        return (m.description, m.amount, m.released);
    }
}
