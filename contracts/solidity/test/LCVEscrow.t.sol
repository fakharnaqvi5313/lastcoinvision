// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import {Test} from "forge-std/Test.sol";
import {LCVEscrow} from "../src/LCVEscrow.sol";

contract LCVEscrowTest is Test {
    LCVEscrow escrow;

    address buyer = makeAddr("buyer");
    address seller = makeAddr("seller");
    address arbiter = makeAddr("arbiter");
    address stranger = makeAddr("stranger");

    function setUp() public {
        escrow = new LCVEscrow();
        vm.deal(buyer, 100 ether);
    }

    function _createStandardEscrow() internal returns (uint256 id) {
        string[] memory descriptions = new string[](3);
        descriptions[0] = "Foundation complete";
        descriptions[1] = "Structure complete";
        descriptions[2] = "Handover";
        uint256[] memory amounts = new uint256[](3);
        amounts[0] = 3 ether;
        amounts[1] = 4 ether;
        amounts[2] = 3 ether;

        vm.prank(buyer);
        id = escrow.createEscrow(seller, arbiter, descriptions, amounts);
    }

    function test_CreateEscrow_SetsTotalAmount() public {
        uint256 id = _createStandardEscrow();
        (address b, address s, address a, uint256 total,,,, uint256 milestoneCount) = escrow.getEscrow(id);
        assertEq(b, buyer);
        assertEq(s, seller);
        assertEq(a, arbiter);
        assertEq(total, 10 ether);
        assertEq(milestoneCount, 3);
    }

    function test_Deposit_AccumulatesUpToTotal() public {
        uint256 id = _createStandardEscrow();

        vm.prank(buyer);
        escrow.deposit{value: 6 ether}(id);
        vm.prank(buyer);
        escrow.deposit{value: 4 ether}(id);

        (,,,, uint256 deposited,,,) = escrow.getEscrow(id);
        assertEq(deposited, 10 ether);
    }

    function test_Deposit_RevertsIfExceedsTotal() public {
        uint256 id = _createStandardEscrow();
        vm.prank(buyer);
        vm.expectRevert(LCVEscrow.AmountMismatch.selector);
        escrow.deposit{value: 11 ether}(id);
    }

    function test_Deposit_RevertsIfNotBuyer() public {
        uint256 id = _createStandardEscrow();
        vm.deal(stranger, 1 ether);
        vm.prank(stranger);
        vm.expectRevert(LCVEscrow.NotBuyer.selector);
        escrow.deposit{value: 1 ether}(id);
    }

    function test_ReleaseMilestone_RequiresFullDeposit() public {
        uint256 id = _createStandardEscrow();
        vm.prank(buyer);
        escrow.deposit{value: 5 ether}(id);

        vm.prank(arbiter);
        vm.expectRevert(LCVEscrow.NotFullyDeposited.selector);
        escrow.releaseMilestone(id, 0);
    }

    function test_ReleaseMilestone_PaysSellerAndTracksState() public {
        uint256 id = _createStandardEscrow();
        vm.prank(buyer);
        escrow.deposit{value: 10 ether}(id);

        uint256 sellerBalanceBefore = seller.balance;

        vm.prank(arbiter);
        escrow.releaseMilestone(id, 0);

        assertEq(seller.balance, sellerBalanceBefore + 3 ether);
        (,,,,, uint256 released,,) = escrow.getEscrow(id);
        assertEq(released, 3 ether);

        (,, bool wasReleased) = escrow.getMilestone(id, 0);
        assertTrue(wasReleased);
    }

    function test_ReleaseMilestone_RevertsIfAlreadyReleased() public {
        uint256 id = _createStandardEscrow();
        vm.prank(buyer);
        escrow.deposit{value: 10 ether}(id);

        vm.prank(arbiter);
        escrow.releaseMilestone(id, 0);

        vm.prank(arbiter);
        vm.expectRevert(LCVEscrow.MilestoneAlreadyReleased.selector);
        escrow.releaseMilestone(id, 0);
    }

    function test_ReleaseMilestone_RevertsIfNotArbiter() public {
        uint256 id = _createStandardEscrow();
        vm.prank(buyer);
        escrow.deposit{value: 10 ether}(id);

        vm.prank(buyer);
        vm.expectRevert(LCVEscrow.NotArbiter.selector);
        escrow.releaseMilestone(id, 0);
    }

    function test_CancelAndRefund_ReturnsUndisbursedFundsToBuyer() public {
        uint256 id = _createStandardEscrow();
        vm.prank(buyer);
        escrow.deposit{value: 10 ether}(id);

        vm.prank(arbiter);
        escrow.releaseMilestone(id, 0); // 3 ether released, 7 ether remains

        uint256 buyerBalanceBefore = buyer.balance;

        vm.prank(arbiter);
        escrow.cancelAndRefund(id);

        assertEq(buyer.balance, buyerBalanceBefore + 7 ether);

        (,,,,,, bool cancelled,) = escrow.getEscrow(id);
        assertTrue(cancelled);
    }

    function test_CancelAndRefund_RevertsIfNotArbiter() public {
        uint256 id = _createStandardEscrow();
        vm.prank(buyer);
        escrow.deposit{value: 10 ether}(id);

        vm.prank(seller);
        vm.expectRevert(LCVEscrow.NotArbiter.selector);
        escrow.cancelAndRefund(id);
    }

    function test_ReleaseMilestone_RevertsAfterCancellation() public {
        uint256 id = _createStandardEscrow();
        vm.prank(buyer);
        escrow.deposit{value: 10 ether}(id);

        vm.prank(arbiter);
        escrow.cancelAndRefund(id);

        vm.prank(arbiter);
        vm.expectRevert(LCVEscrow.EscrowIsCancelled.selector);
        escrow.releaseMilestone(id, 1);
    }

    function test_CreateEscrow_RevertsOnMismatchedArrays() public {
        string[] memory descriptions = new string[](2);
        descriptions[0] = "a";
        descriptions[1] = "b";
        uint256[] memory amounts = new uint256[](1);
        amounts[0] = 1 ether;

        vm.prank(buyer);
        vm.expectRevert(LCVEscrow.NoMilestones.selector);
        escrow.createEscrow(seller, arbiter, descriptions, amounts);
    }
}
