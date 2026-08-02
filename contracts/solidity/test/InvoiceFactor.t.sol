// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import {Test} from "forge-std/Test.sol";
import {InvoiceFactor} from "../src/InvoiceFactor.sol";

contract InvoiceFactorTest is Test {
    InvoiceFactor factoring;

    address seller = makeAddr("seller");
    address debtor = makeAddr("debtor");
    address factor = makeAddr("factor");
    address stranger = makeAddr("stranger");

    uint256 constant FACE_VALUE = 10 ether;
    uint256 constant ADVANCE = 8 ether;
    uint256 constant FEE = 0.5 ether;

    function setUp() public {
        factoring = new InvoiceFactor();
        vm.deal(factor, 100 ether);
        vm.deal(debtor, 100 ether);
    }

    function _createInvoice() internal returns (uint256 id) {
        vm.prank(seller);
        id = factoring.createInvoice(debtor, FACE_VALUE, ADVANCE, FEE, block.timestamp + 30 days);
    }

    function test_CreateInvoice_RevertsIfAdvancePlusFeeExceedsFaceValue() public {
        vm.prank(seller);
        vm.expectRevert(InvoiceFactor.AdvancePlusFeeExceedsFaceValue.selector);
        factoring.createInvoice(debtor, 10 ether, 9 ether, 2 ether, block.timestamp + 30 days);
    }

    function test_CreateInvoice_RevertsIfDueDateInPast() public {
        vm.warp(1000);
        vm.prank(seller);
        vm.expectRevert(InvoiceFactor.DueDateInPast.selector);
        factoring.createInvoice(debtor, FACE_VALUE, ADVANCE, FEE, 500);
    }

    function test_FundInvoice_PaysSellerImmediately() public {
        uint256 id = _createInvoice();
        uint256 sellerBalanceBefore = seller.balance;

        vm.prank(factor);
        factoring.fundInvoice{value: ADVANCE}(id);

        assertEq(seller.balance, sellerBalanceBefore + ADVANCE);

        (,, address recordedFactor,,,,, InvoiceFactor.Status status) = factoring.invoices(id);
        assertEq(recordedFactor, factor);
        assertEq(uint8(status), uint8(InvoiceFactor.Status.Funded));
    }

    function test_FundInvoice_RevertsOnWrongAmount() public {
        uint256 id = _createInvoice();
        vm.prank(factor);
        vm.expectRevert(InvoiceFactor.WrongAmount.selector);
        factoring.fundInvoice{value: ADVANCE - 1}(id);
    }

    function test_FundInvoice_RevertsIfAlreadyFunded() public {
        uint256 id = _createInvoice();
        vm.prank(factor);
        factoring.fundInvoice{value: ADVANCE}(id);

        vm.prank(stranger);
        vm.deal(stranger, ADVANCE);
        vm.expectRevert(InvoiceFactor.WrongStatus.selector);
        factoring.fundInvoice{value: ADVANCE}(id);
    }

    function test_PayInvoice_RevertsIfNotDebtor() public {
        uint256 id = _createInvoice();
        vm.prank(factor);
        factoring.fundInvoice{value: ADVANCE}(id);

        vm.deal(stranger, FACE_VALUE);
        vm.prank(stranger);
        vm.expectRevert(InvoiceFactor.NotDebtor.selector);
        factoring.payInvoice{value: FACE_VALUE}(id);
    }

    function test_FullLifecycle_SettlesFactorAndSellerCorrectly() public {
        uint256 id = _createInvoice();

        vm.prank(factor);
        factoring.fundInvoice{value: ADVANCE}(id);

        vm.prank(debtor);
        factoring.payInvoice{value: FACE_VALUE}(id);

        uint256 factorBalanceBefore = factor.balance;
        uint256 sellerBalanceBefore = seller.balance;

        factoring.settle(id);

        assertEq(factor.balance, factorBalanceBefore + ADVANCE + FEE);
        assertEq(seller.balance, sellerBalanceBefore + (FACE_VALUE - ADVANCE - FEE));

        (,,,,,,, InvoiceFactor.Status status) = factoring.invoices(id);
        assertEq(uint8(status), uint8(InvoiceFactor.Status.Settled));
    }

    function test_Settle_RevertsIfNotYetPaid() public {
        uint256 id = _createInvoice();
        vm.prank(factor);
        factoring.fundInvoice{value: ADVANCE}(id);

        vm.expectRevert(InvoiceFactor.WrongStatus.selector);
        factoring.settle(id);
    }

    function test_CancelInvoice_OnlyBeforeFunding() public {
        uint256 id = _createInvoice();

        vm.prank(seller);
        factoring.cancelInvoice(id);

        (,,,,,,, InvoiceFactor.Status status) = factoring.invoices(id);
        assertEq(uint8(status), uint8(InvoiceFactor.Status.Cancelled));
    }

    function test_CancelInvoice_RevertsAfterFunding() public {
        uint256 id = _createInvoice();
        vm.prank(factor);
        factoring.fundInvoice{value: ADVANCE}(id);

        vm.prank(seller);
        vm.expectRevert(InvoiceFactor.WrongStatus.selector);
        factoring.cancelInvoice(id);
    }

    function test_CancelInvoice_RevertsIfNotSeller() public {
        uint256 id = _createInvoice();
        vm.prank(stranger);
        vm.expectRevert(InvoiceFactor.NotSeller.selector);
        factoring.cancelInvoice(id);
    }
}
