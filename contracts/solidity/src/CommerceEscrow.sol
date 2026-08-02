// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

/// @title CommerceEscrow
/// @notice Single-shot buyer-seller escrow with an arbiter for general commerce (goods
/// purchases, freelance deliverables, etc). The buyer deposits payment up front; the seller
/// delivers off-chain; the buyer confirms receipt to release funds. If the buyer never
/// confirms, the seller can claim after a confirmation window elapses. Either party may raise
/// a dispute before then, which only the arbiter can resolve.
/// @dev Reference contract for LastcoinVision's EVM compatibility layer (Phase D). Native-token
/// (LCV) denominated.
contract CommerceEscrow {
    enum Status {
        AwaitingDelivery,
        Disputed,
        Complete,
        Refunded
    }

    struct Deal {
        address buyer;
        address seller;
        address arbiter;
        uint256 amount;
        uint256 confirmDeadline;
        Status status;
    }

    uint256 public nextDealId;
    mapping(uint256 => Deal) public deals;

    event DealCreated(uint256 indexed dealId, address indexed buyer, address indexed seller, address arbiter, uint256 amount, uint256 confirmDeadline);
    event ReceiptConfirmed(uint256 indexed dealId);
    event ClaimedAfterTimeout(uint256 indexed dealId);
    event DisputeRaised(uint256 indexed dealId, address indexed raisedBy);
    event DisputeResolved(uint256 indexed dealId, bool releasedToSeller);

    error ZeroAddress();
    error ZeroAmount();
    error NotBuyer();
    error NotSeller();
    error NotArbiter();
    error NotParticipant();
    error WrongStatus();
    error DeadlineNotReached();

    modifier onlyParticipant(uint256 dealId) {
        Deal storage d = deals[dealId];
        if (msg.sender != d.buyer && msg.sender != d.seller) revert NotParticipant();
        _;
    }

    /// @notice The buyer opens a deal and deposits the full payment in the same transaction.
    /// `confirmWindow` is how long, in seconds from now, the seller must wait before being
    /// able to claim payment automatically if the buyer never confirms receipt.
    function createDeal(address seller, address arbiter, uint256 confirmWindow) external payable returns (uint256 dealId) {
        if (seller == address(0) || arbiter == address(0)) revert ZeroAddress();
        if (msg.value == 0) revert ZeroAmount();

        dealId = nextDealId++;
        uint256 deadline = block.timestamp + confirmWindow;
        deals[dealId] = Deal({
            buyer: msg.sender,
            seller: seller,
            arbiter: arbiter,
            amount: msg.value,
            confirmDeadline: deadline,
            status: Status.AwaitingDelivery
        });

        emit DealCreated(dealId, msg.sender, seller, arbiter, msg.value, deadline);
    }

    /// @notice The buyer confirms delivery was received as expected, releasing funds to the
    /// seller immediately.
    function confirmReceipt(uint256 dealId) external {
        Deal storage d = deals[dealId];
        if (msg.sender != d.buyer) revert NotBuyer();
        if (d.status != Status.AwaitingDelivery) revert WrongStatus();

        d.status = Status.Complete;
        emit ReceiptConfirmed(dealId);

        (bool ok, ) = d.seller.call{value: d.amount}("");
        require(ok, "transfer to seller failed");
    }

    /// @notice If the buyer never confirms and never disputes, the seller may claim payment
    /// once the confirmation window has elapsed.
    function claimAfterTimeout(uint256 dealId) external {
        Deal storage d = deals[dealId];
        if (msg.sender != d.seller) revert NotSeller();
        if (d.status != Status.AwaitingDelivery) revert WrongStatus();
        if (block.timestamp < d.confirmDeadline) revert DeadlineNotReached();

        d.status = Status.Complete;
        emit ClaimedAfterTimeout(dealId);

        (bool ok, ) = d.seller.call{value: d.amount}("");
        require(ok, "transfer to seller failed");
    }

    /// @notice Either the buyer or the seller may raise a dispute before the deal is otherwise
    /// settled, freezing it until the arbiter resolves it.
    function raiseDispute(uint256 dealId) external onlyParticipant(dealId) {
        Deal storage d = deals[dealId];
        if (d.status != Status.AwaitingDelivery) revert WrongStatus();

        d.status = Status.Disputed;
        emit DisputeRaised(dealId, msg.sender);
    }

    /// @notice The arbiter resolves a dispute by releasing the deposited funds to either the
    /// seller or back to the buyer.
    function resolveDispute(uint256 dealId, bool releaseToSeller) external {
        Deal storage d = deals[dealId];
        if (msg.sender != d.arbiter) revert NotArbiter();
        if (d.status != Status.Disputed) revert WrongStatus();

        address recipient;
        if (releaseToSeller) {
            d.status = Status.Complete;
            recipient = d.seller;
        } else {
            d.status = Status.Refunded;
            recipient = d.buyer;
        }
        emit DisputeResolved(dealId, releaseToSeller);

        (bool ok, ) = recipient.call{value: d.amount}("");
        require(ok, "settlement transfer failed");
    }
}
