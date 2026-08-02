// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import {Test} from "forge-std/Test.sol";
import {CommerceEscrow} from "../src/CommerceEscrow.sol";

contract CommerceEscrowTest is Test {
    CommerceEscrow escrow;

    address buyer = makeAddr("buyer");
    address seller = makeAddr("seller");
    address arbiter = makeAddr("arbiter");
    address stranger = makeAddr("stranger");

    uint256 constant AMOUNT = 5 ether;
    uint256 constant WINDOW = 7 days;

    function setUp() public {
        escrow = new CommerceEscrow();
        vm.deal(buyer, 100 ether);
    }

    function _createDeal() internal returns (uint256 id) {
        vm.prank(buyer);
        id = escrow.createDeal{value: AMOUNT}(seller, arbiter, WINDOW);
    }

    function test_CreateDeal_RevertsOnZeroAmount() public {
        vm.prank(buyer);
        vm.expectRevert(CommerceEscrow.ZeroAmount.selector);
        escrow.createDeal{value: 0}(seller, arbiter, WINDOW);
    }

    function test_ConfirmReceipt_ReleasesToSeller() public {
        uint256 id = _createDeal();
        uint256 sellerBalanceBefore = seller.balance;

        vm.prank(buyer);
        escrow.confirmReceipt(id);

        assertEq(seller.balance, sellerBalanceBefore + AMOUNT);
        (,,,,, CommerceEscrow.Status status) = escrow.deals(id);
        assertEq(uint8(status), uint8(CommerceEscrow.Status.Complete));
    }

    function test_ConfirmReceipt_RevertsIfNotBuyer() public {
        uint256 id = _createDeal();
        vm.prank(stranger);
        vm.expectRevert(CommerceEscrow.NotBuyer.selector);
        escrow.confirmReceipt(id);
    }

    function test_ConfirmReceipt_RevertsIfAlreadyComplete() public {
        uint256 id = _createDeal();
        vm.prank(buyer);
        escrow.confirmReceipt(id);

        vm.prank(buyer);
        vm.expectRevert(CommerceEscrow.WrongStatus.selector);
        escrow.confirmReceipt(id);
    }

    function test_ClaimAfterTimeout_RevertsBeforeDeadline() public {
        uint256 id = _createDeal();
        vm.prank(seller);
        vm.expectRevert(CommerceEscrow.DeadlineNotReached.selector);
        escrow.claimAfterTimeout(id);
    }

    function test_ClaimAfterTimeout_PaysSellerAfterDeadline() public {
        uint256 id = _createDeal();
        vm.warp(block.timestamp + WINDOW + 1);

        uint256 sellerBalanceBefore = seller.balance;
        vm.prank(seller);
        escrow.claimAfterTimeout(id);

        assertEq(seller.balance, sellerBalanceBefore + AMOUNT);
    }

    function test_ClaimAfterTimeout_RevertsIfNotSeller() public {
        uint256 id = _createDeal();
        vm.warp(block.timestamp + WINDOW + 1);

        vm.prank(stranger);
        vm.expectRevert(CommerceEscrow.NotSeller.selector);
        escrow.claimAfterTimeout(id);
    }

    function test_RaiseDispute_RevertsIfNotParticipant() public {
        uint256 id = _createDeal();
        vm.prank(stranger);
        vm.expectRevert(CommerceEscrow.NotParticipant.selector);
        escrow.raiseDispute(id);
    }

    function test_DisputeResolution_ReleaseToSeller() public {
        uint256 id = _createDeal();
        vm.prank(buyer);
        escrow.raiseDispute(id);

        uint256 sellerBalanceBefore = seller.balance;
        vm.prank(arbiter);
        escrow.resolveDispute(id, true);

        assertEq(seller.balance, sellerBalanceBefore + AMOUNT);
        (,,,,, CommerceEscrow.Status status) = escrow.deals(id);
        assertEq(uint8(status), uint8(CommerceEscrow.Status.Complete));
    }

    function test_DisputeResolution_RefundToBuyer() public {
        uint256 id = _createDeal();
        vm.prank(seller);
        escrow.raiseDispute(id);

        uint256 buyerBalanceBefore = buyer.balance;
        vm.prank(arbiter);
        escrow.resolveDispute(id, false);

        assertEq(buyer.balance, buyerBalanceBefore + AMOUNT);
        (,,,,, CommerceEscrow.Status status) = escrow.deals(id);
        assertEq(uint8(status), uint8(CommerceEscrow.Status.Refunded));
    }

    function test_ResolveDispute_RevertsIfNotArbiter() public {
        uint256 id = _createDeal();
        vm.prank(buyer);
        escrow.raiseDispute(id);

        vm.prank(seller);
        vm.expectRevert(CommerceEscrow.NotArbiter.selector);
        escrow.resolveDispute(id, true);
    }

    function test_ClaimAfterTimeout_RevertsOnceDisputed() public {
        uint256 id = _createDeal();
        vm.prank(buyer);
        escrow.raiseDispute(id);

        vm.warp(block.timestamp + WINDOW + 1);
        vm.prank(seller);
        vm.expectRevert(CommerceEscrow.WrongStatus.selector);
        escrow.claimAfterTimeout(id);
    }
}
