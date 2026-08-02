// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

/// @title InvoiceFactor
/// @notice Multi-party invoice financing (factoring). A seller holding a receivable from a
/// debtor lists it for financing; a factor advances the seller cash immediately in exchange
/// for the right to collect the invoice's face value (plus a fee) when the debtor eventually
/// pays. Non-recourse: if the debtor never pays, the factor bears the loss and the seller keeps
/// the advance — the factor is expected to have already priced that risk into the discount.
/// @dev Reference contract for LastcoinVision's EVM compatibility layer (Phase D). Native-token
/// (LCV) denominated.
contract InvoiceFactor {
    enum Status {
        Created,
        Funded,
        Paid,
        Settled,
        Cancelled
    }

    struct Invoice {
        address seller;
        address debtor;
        address factor;
        uint256 faceValue;
        uint256 advanceAmount;
        uint256 factorFee;
        uint256 dueDate;
        Status status;
    }

    uint256 public nextInvoiceId;
    mapping(uint256 => Invoice) public invoices;

    event InvoiceCreated(uint256 indexed invoiceId, address indexed seller, address indexed debtor, uint256 faceValue, uint256 advanceAmount, uint256 factorFee, uint256 dueDate);
    event InvoiceFunded(uint256 indexed invoiceId, address indexed factor, uint256 advanceAmount);
    event InvoicePaid(uint256 indexed invoiceId, address indexed debtor, uint256 amount);
    event InvoiceSettled(uint256 indexed invoiceId, uint256 factorPayout, uint256 sellerRebate);
    event InvoiceCancelled(uint256 indexed invoiceId);

    error ZeroAddress();
    error ZeroFaceValue();
    error AdvancePlusFeeExceedsFaceValue();
    error DueDateInPast();
    error NotSeller();
    error NotDebtor();
    error WrongStatus();
    error WrongAmount();

    modifier onlySeller(uint256 invoiceId) {
        if (msg.sender != invoices[invoiceId].seller) revert NotSeller();
        _;
    }

    /// @notice Lists a new invoice for factoring.
    function createInvoice(
        address debtor,
        uint256 faceValue,
        uint256 advanceAmount,
        uint256 factorFee,
        uint256 dueDate
    ) external returns (uint256 invoiceId) {
        if (debtor == address(0)) revert ZeroAddress();
        if (faceValue == 0) revert ZeroFaceValue();
        if (advanceAmount + factorFee > faceValue) revert AdvancePlusFeeExceedsFaceValue();
        if (dueDate <= block.timestamp) revert DueDateInPast();

        invoiceId = nextInvoiceId++;
        invoices[invoiceId] = Invoice({
            seller: msg.sender,
            debtor: debtor,
            factor: address(0),
            faceValue: faceValue,
            advanceAmount: advanceAmount,
            factorFee: factorFee,
            dueDate: dueDate,
            status: Status.Created
        });

        emit InvoiceCreated(invoiceId, msg.sender, debtor, faceValue, advanceAmount, factorFee, dueDate);
    }

    /// @notice A factor funds the invoice, paying the advance amount straight to the seller.
    function fundInvoice(uint256 invoiceId) external payable {
        Invoice storage inv = invoices[invoiceId];
        if (inv.status != Status.Created) revert WrongStatus();
        if (msg.value != inv.advanceAmount) revert WrongAmount();

        inv.factor = msg.sender;
        inv.status = Status.Funded;

        emit InvoiceFunded(invoiceId, msg.sender, msg.value);
        (bool ok, ) = inv.seller.call{value: msg.value}("");
        require(ok, "advance transfer to seller failed");
    }

    /// @notice The debtor pays the invoice's full face value into the contract, to be
    /// distributed between the factor and the seller via `settle`.
    function payInvoice(uint256 invoiceId) external payable {
        Invoice storage inv = invoices[invoiceId];
        if (inv.status != Status.Funded) revert WrongStatus();
        if (msg.sender != inv.debtor) revert NotDebtor();
        if (msg.value != inv.faceValue) revert WrongAmount();

        inv.status = Status.Paid;
        emit InvoicePaid(invoiceId, msg.sender, msg.value);
    }

    /// @notice Distributes a paid invoice: the factor recovers its advance plus its fee, and
    /// the seller receives whatever remains of the face value.
    function settle(uint256 invoiceId) external {
        Invoice storage inv = invoices[invoiceId];
        if (inv.status != Status.Paid) revert WrongStatus();

        inv.status = Status.Settled;

        uint256 factorPayout = inv.advanceAmount + inv.factorFee;
        uint256 sellerRebate = inv.faceValue - factorPayout;

        emit InvoiceSettled(invoiceId, factorPayout, sellerRebate);

        if (factorPayout > 0) {
            (bool okFactor, ) = inv.factor.call{value: factorPayout}("");
            require(okFactor, "payout to factor failed");
        }
        if (sellerRebate > 0) {
            (bool okSeller, ) = inv.seller.call{value: sellerRebate}("");
            require(okSeller, "rebate to seller failed");
        }
    }

    /// @notice The seller may cancel a listing that has not yet been funded by a factor.
    function cancelInvoice(uint256 invoiceId) external onlySeller(invoiceId) {
        Invoice storage inv = invoices[invoiceId];
        if (inv.status != Status.Created) revert WrongStatus();

        inv.status = Status.Cancelled;
        emit InvoiceCancelled(invoiceId);
    }
}
