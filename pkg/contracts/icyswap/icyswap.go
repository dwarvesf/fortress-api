// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package icyswap

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// IcySwapMetaData contains all meta data concerning the IcySwap contract.
var IcySwapMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"_usdc\",\"type\":\"address\"},{\"internalType\":\"contractIERC20\",\"name\":\"_icy\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_conversionRate\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"conversionRate\",\"type\":\"uint256\"}],\"name\":\"ConversionRateChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"Paused\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"contractIERC20\",\"name\":\"fromToken\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"fromAmount\",\"type\":\"uint256\"}],\"name\":\"Swap\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"Unpaused\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"contractIERC20\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"WithdrawToOwner\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"icy\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"icyToUsdcConversionRate\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pause\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"paused\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_conversionRate\",\"type\":\"uint256\"}],\"name\":\"setConversionRate\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amountIn\",\"type\":\"uint256\"}],\"name\":\"swap\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"unpause\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"usdc\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"withdrawToOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// IcySwapABI is the input ABI used to generate the binding from.
// Deprecated: Use IcySwapMetaData.ABI instead.
var IcySwapABI = IcySwapMetaData.ABI

// IcySwap is an auto generated Go binding around an Ethereum contract.
type IcySwap struct {
	IcySwapCaller     // Read-only binding to the contract
	IcySwapTransactor // Write-only binding to the contract
	IcySwapFilterer   // Log filterer for contract events
}

// IcySwapCaller is an auto generated read-only Go binding around an Ethereum contract.
type IcySwapCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IcySwapTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IcySwapTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IcySwapFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IcySwapFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IcySwapSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IcySwapSession struct {
	Contract     *IcySwap          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IcySwapCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IcySwapCallerSession struct {
	Contract *IcySwapCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// IcySwapTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IcySwapTransactorSession struct {
	Contract     *IcySwapTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// IcySwapRaw is an auto generated low-level Go binding around an Ethereum contract.
type IcySwapRaw struct {
	Contract *IcySwap // Generic contract binding to access the raw methods on
}

// IcySwapCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IcySwapCallerRaw struct {
	Contract *IcySwapCaller // Generic read-only contract binding to access the raw methods on
}

// IcySwapTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IcySwapTransactorRaw struct {
	Contract *IcySwapTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIcySwap creates a new instance of IcySwap, bound to a specific deployed contract.
func NewIcySwap(address common.Address, backend bind.ContractBackend) (*IcySwap, error) {
	contract, err := bindIcySwap(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IcySwap{IcySwapCaller: IcySwapCaller{contract: contract}, IcySwapTransactor: IcySwapTransactor{contract: contract}, IcySwapFilterer: IcySwapFilterer{contract: contract}}, nil
}

// NewIcySwapCaller creates a new read-only instance of IcySwap, bound to a specific deployed contract.
func NewIcySwapCaller(address common.Address, caller bind.ContractCaller) (*IcySwapCaller, error) {
	contract, err := bindIcySwap(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IcySwapCaller{contract: contract}, nil
}

// NewIcySwapTransactor creates a new write-only instance of IcySwap, bound to a specific deployed contract.
func NewIcySwapTransactor(address common.Address, transactor bind.ContractTransactor) (*IcySwapTransactor, error) {
	contract, err := bindIcySwap(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IcySwapTransactor{contract: contract}, nil
}

// NewIcySwapFilterer creates a new log filterer instance of IcySwap, bound to a specific deployed contract.
func NewIcySwapFilterer(address common.Address, filterer bind.ContractFilterer) (*IcySwapFilterer, error) {
	contract, err := bindIcySwap(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IcySwapFilterer{contract: contract}, nil
}

// bindIcySwap binds a generic wrapper to an already deployed contract.
func bindIcySwap(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IcySwapABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IcySwap *IcySwapRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IcySwap.Contract.IcySwapCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IcySwap *IcySwapRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IcySwap.Contract.IcySwapTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IcySwap *IcySwapRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IcySwap.Contract.IcySwapTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IcySwap *IcySwapCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IcySwap.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IcySwap *IcySwapTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IcySwap.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IcySwap *IcySwapTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IcySwap.Contract.contract.Transact(opts, method, params...)
}

// Icy is a free data retrieval call binding the contract method 0x7f245ab1.
//
// Solidity: function icy() view returns(address)
func (_IcySwap *IcySwapCaller) Icy(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _IcySwap.contract.Call(opts, &out, "icy")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Icy is a free data retrieval call binding the contract method 0x7f245ab1.
//
// Solidity: function icy() view returns(address)
func (_IcySwap *IcySwapSession) Icy() (common.Address, error) {
	return _IcySwap.Contract.Icy(&_IcySwap.CallOpts)
}

// Icy is a free data retrieval call binding the contract method 0x7f245ab1.
//
// Solidity: function icy() view returns(address)
func (_IcySwap *IcySwapCallerSession) Icy() (common.Address, error) {
	return _IcySwap.Contract.Icy(&_IcySwap.CallOpts)
}

// IcyToUsdcConversionRate is a free data retrieval call binding the contract method 0x0e7469f3.
//
// Solidity: function icyToUsdcConversionRate() view returns(uint256)
func (_IcySwap *IcySwapCaller) IcyToUsdcConversionRate(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _IcySwap.contract.Call(opts, &out, "icyToUsdcConversionRate")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// IcyToUsdcConversionRate is a free data retrieval call binding the contract method 0x0e7469f3.
//
// Solidity: function icyToUsdcConversionRate() view returns(uint256)
func (_IcySwap *IcySwapSession) IcyToUsdcConversionRate() (*big.Int, error) {
	return _IcySwap.Contract.IcyToUsdcConversionRate(&_IcySwap.CallOpts)
}

// IcyToUsdcConversionRate is a free data retrieval call binding the contract method 0x0e7469f3.
//
// Solidity: function icyToUsdcConversionRate() view returns(uint256)
func (_IcySwap *IcySwapCallerSession) IcyToUsdcConversionRate() (*big.Int, error) {
	return _IcySwap.Contract.IcyToUsdcConversionRate(&_IcySwap.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_IcySwap *IcySwapCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _IcySwap.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_IcySwap *IcySwapSession) Owner() (common.Address, error) {
	return _IcySwap.Contract.Owner(&_IcySwap.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_IcySwap *IcySwapCallerSession) Owner() (common.Address, error) {
	return _IcySwap.Contract.Owner(&_IcySwap.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_IcySwap *IcySwapCaller) Paused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _IcySwap.contract.Call(opts, &out, "paused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_IcySwap *IcySwapSession) Paused() (bool, error) {
	return _IcySwap.Contract.Paused(&_IcySwap.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_IcySwap *IcySwapCallerSession) Paused() (bool, error) {
	return _IcySwap.Contract.Paused(&_IcySwap.CallOpts)
}

// Usdc is a free data retrieval call binding the contract method 0x3e413bee.
//
// Solidity: function usdc() view returns(address)
func (_IcySwap *IcySwapCaller) Usdc(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _IcySwap.contract.Call(opts, &out, "usdc")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Usdc is a free data retrieval call binding the contract method 0x3e413bee.
//
// Solidity: function usdc() view returns(address)
func (_IcySwap *IcySwapSession) Usdc() (common.Address, error) {
	return _IcySwap.Contract.Usdc(&_IcySwap.CallOpts)
}

// Usdc is a free data retrieval call binding the contract method 0x3e413bee.
//
// Solidity: function usdc() view returns(address)
func (_IcySwap *IcySwapCallerSession) Usdc() (common.Address, error) {
	return _IcySwap.Contract.Usdc(&_IcySwap.CallOpts)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_IcySwap *IcySwapTransactor) Pause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IcySwap.contract.Transact(opts, "pause")
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_IcySwap *IcySwapSession) Pause() (*types.Transaction, error) {
	return _IcySwap.Contract.Pause(&_IcySwap.TransactOpts)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_IcySwap *IcySwapTransactorSession) Pause() (*types.Transaction, error) {
	return _IcySwap.Contract.Pause(&_IcySwap.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_IcySwap *IcySwapTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IcySwap.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_IcySwap *IcySwapSession) RenounceOwnership() (*types.Transaction, error) {
	return _IcySwap.Contract.RenounceOwnership(&_IcySwap.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_IcySwap *IcySwapTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _IcySwap.Contract.RenounceOwnership(&_IcySwap.TransactOpts)
}

// SetConversionRate is a paid mutator transaction binding the contract method 0xd2e80494.
//
// Solidity: function setConversionRate(uint256 _conversionRate) returns()
func (_IcySwap *IcySwapTransactor) SetConversionRate(opts *bind.TransactOpts, _conversionRate *big.Int) (*types.Transaction, error) {
	return _IcySwap.contract.Transact(opts, "setConversionRate", _conversionRate)
}

// SetConversionRate is a paid mutator transaction binding the contract method 0xd2e80494.
//
// Solidity: function setConversionRate(uint256 _conversionRate) returns()
func (_IcySwap *IcySwapSession) SetConversionRate(_conversionRate *big.Int) (*types.Transaction, error) {
	return _IcySwap.Contract.SetConversionRate(&_IcySwap.TransactOpts, _conversionRate)
}

// SetConversionRate is a paid mutator transaction binding the contract method 0xd2e80494.
//
// Solidity: function setConversionRate(uint256 _conversionRate) returns()
func (_IcySwap *IcySwapTransactorSession) SetConversionRate(_conversionRate *big.Int) (*types.Transaction, error) {
	return _IcySwap.Contract.SetConversionRate(&_IcySwap.TransactOpts, _conversionRate)
}

// Swap is a paid mutator transaction binding the contract method 0x94b918de.
//
// Solidity: function swap(uint256 _amountIn) returns()
func (_IcySwap *IcySwapTransactor) Swap(opts *bind.TransactOpts, _amountIn *big.Int) (*types.Transaction, error) {
	return _IcySwap.contract.Transact(opts, "swap", _amountIn)
}

// Swap is a paid mutator transaction binding the contract method 0x94b918de.
//
// Solidity: function swap(uint256 _amountIn) returns()
func (_IcySwap *IcySwapSession) Swap(_amountIn *big.Int) (*types.Transaction, error) {
	return _IcySwap.Contract.Swap(&_IcySwap.TransactOpts, _amountIn)
}

// Swap is a paid mutator transaction binding the contract method 0x94b918de.
//
// Solidity: function swap(uint256 _amountIn) returns()
func (_IcySwap *IcySwapTransactorSession) Swap(_amountIn *big.Int) (*types.Transaction, error) {
	return _IcySwap.Contract.Swap(&_IcySwap.TransactOpts, _amountIn)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_IcySwap *IcySwapTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _IcySwap.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_IcySwap *IcySwapSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _IcySwap.Contract.TransferOwnership(&_IcySwap.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_IcySwap *IcySwapTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _IcySwap.Contract.TransferOwnership(&_IcySwap.TransactOpts, newOwner)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_IcySwap *IcySwapTransactor) Unpause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IcySwap.contract.Transact(opts, "unpause")
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_IcySwap *IcySwapSession) Unpause() (*types.Transaction, error) {
	return _IcySwap.Contract.Unpause(&_IcySwap.TransactOpts)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_IcySwap *IcySwapTransactorSession) Unpause() (*types.Transaction, error) {
	return _IcySwap.Contract.Unpause(&_IcySwap.TransactOpts)
}

// WithdrawToOwner is a paid mutator transaction binding the contract method 0x360b9cec.
//
// Solidity: function withdrawToOwner(address _token) returns()
func (_IcySwap *IcySwapTransactor) WithdrawToOwner(opts *bind.TransactOpts, _token common.Address) (*types.Transaction, error) {
	return _IcySwap.contract.Transact(opts, "withdrawToOwner", _token)
}

// WithdrawToOwner is a paid mutator transaction binding the contract method 0x360b9cec.
//
// Solidity: function withdrawToOwner(address _token) returns()
func (_IcySwap *IcySwapSession) WithdrawToOwner(_token common.Address) (*types.Transaction, error) {
	return _IcySwap.Contract.WithdrawToOwner(&_IcySwap.TransactOpts, _token)
}

// WithdrawToOwner is a paid mutator transaction binding the contract method 0x360b9cec.
//
// Solidity: function withdrawToOwner(address _token) returns()
func (_IcySwap *IcySwapTransactorSession) WithdrawToOwner(_token common.Address) (*types.Transaction, error) {
	return _IcySwap.Contract.WithdrawToOwner(&_IcySwap.TransactOpts, _token)
}

// IcySwapConversionRateChangedIterator is returned from FilterConversionRateChanged and is used to iterate over the raw logs and unpacked data for ConversionRateChanged events raised by the IcySwap contract.
type IcySwapConversionRateChangedIterator struct {
	Event *IcySwapConversionRateChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *IcySwapConversionRateChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IcySwapConversionRateChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(IcySwapConversionRateChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *IcySwapConversionRateChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IcySwapConversionRateChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IcySwapConversionRateChanged represents a ConversionRateChanged event raised by the IcySwap contract.
type IcySwapConversionRateChanged struct {
	ConversionRate *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterConversionRateChanged is a free log retrieval operation binding the contract event 0xb6e78d5c7115b12c1603fb3c8926acd812db2d83d01f62004c460b33f62a8864.
//
// Solidity: event ConversionRateChanged(uint256 conversionRate)
func (_IcySwap *IcySwapFilterer) FilterConversionRateChanged(opts *bind.FilterOpts) (*IcySwapConversionRateChangedIterator, error) {

	logs, sub, err := _IcySwap.contract.FilterLogs(opts, "ConversionRateChanged")
	if err != nil {
		return nil, err
	}
	return &IcySwapConversionRateChangedIterator{contract: _IcySwap.contract, event: "ConversionRateChanged", logs: logs, sub: sub}, nil
}

// WatchConversionRateChanged is a free log subscription operation binding the contract event 0xb6e78d5c7115b12c1603fb3c8926acd812db2d83d01f62004c460b33f62a8864.
//
// Solidity: event ConversionRateChanged(uint256 conversionRate)
func (_IcySwap *IcySwapFilterer) WatchConversionRateChanged(opts *bind.WatchOpts, sink chan<- *IcySwapConversionRateChanged) (event.Subscription, error) {

	logs, sub, err := _IcySwap.contract.WatchLogs(opts, "ConversionRateChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IcySwapConversionRateChanged)
				if err := _IcySwap.contract.UnpackLog(event, "ConversionRateChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseConversionRateChanged is a log parse operation binding the contract event 0xb6e78d5c7115b12c1603fb3c8926acd812db2d83d01f62004c460b33f62a8864.
//
// Solidity: event ConversionRateChanged(uint256 conversionRate)
func (_IcySwap *IcySwapFilterer) ParseConversionRateChanged(log types.Log) (*IcySwapConversionRateChanged, error) {
	event := new(IcySwapConversionRateChanged)
	if err := _IcySwap.contract.UnpackLog(event, "ConversionRateChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IcySwapOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the IcySwap contract.
type IcySwapOwnershipTransferredIterator struct {
	Event *IcySwapOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *IcySwapOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IcySwapOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(IcySwapOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *IcySwapOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IcySwapOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IcySwapOwnershipTransferred represents a OwnershipTransferred event raised by the IcySwap contract.
type IcySwapOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_IcySwap *IcySwapFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*IcySwapOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _IcySwap.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &IcySwapOwnershipTransferredIterator{contract: _IcySwap.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_IcySwap *IcySwapFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *IcySwapOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _IcySwap.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IcySwapOwnershipTransferred)
				if err := _IcySwap.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_IcySwap *IcySwapFilterer) ParseOwnershipTransferred(log types.Log) (*IcySwapOwnershipTransferred, error) {
	event := new(IcySwapOwnershipTransferred)
	if err := _IcySwap.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IcySwapPausedIterator is returned from FilterPaused and is used to iterate over the raw logs and unpacked data for Paused events raised by the IcySwap contract.
type IcySwapPausedIterator struct {
	Event *IcySwapPaused // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *IcySwapPausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IcySwapPaused)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(IcySwapPaused)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *IcySwapPausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IcySwapPausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IcySwapPaused represents a Paused event raised by the IcySwap contract.
type IcySwapPaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterPaused is a free log retrieval operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_IcySwap *IcySwapFilterer) FilterPaused(opts *bind.FilterOpts) (*IcySwapPausedIterator, error) {

	logs, sub, err := _IcySwap.contract.FilterLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return &IcySwapPausedIterator{contract: _IcySwap.contract, event: "Paused", logs: logs, sub: sub}, nil
}

// WatchPaused is a free log subscription operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_IcySwap *IcySwapFilterer) WatchPaused(opts *bind.WatchOpts, sink chan<- *IcySwapPaused) (event.Subscription, error) {

	logs, sub, err := _IcySwap.contract.WatchLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IcySwapPaused)
				if err := _IcySwap.contract.UnpackLog(event, "Paused", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePaused is a log parse operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_IcySwap *IcySwapFilterer) ParsePaused(log types.Log) (*IcySwapPaused, error) {
	event := new(IcySwapPaused)
	if err := _IcySwap.contract.UnpackLog(event, "Paused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IcySwapSwapIterator is returned from FilterSwap and is used to iterate over the raw logs and unpacked data for Swap events raised by the IcySwap contract.
type IcySwapSwapIterator struct {
	Event *IcySwapSwap // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *IcySwapSwapIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IcySwapSwap)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(IcySwapSwap)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *IcySwapSwapIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IcySwapSwapIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IcySwapSwap represents a Swap event raised by the IcySwap contract.
type IcySwapSwap struct {
	FromToken  common.Address
	FromAmount *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterSwap is a free log retrieval operation binding the contract event 0x562c219552544ec4c9d7a8eb850f80ea152973e315372bf4999fe7c953ea004f.
//
// Solidity: event Swap(address indexed fromToken, uint256 indexed fromAmount)
func (_IcySwap *IcySwapFilterer) FilterSwap(opts *bind.FilterOpts, fromToken []common.Address, fromAmount []*big.Int) (*IcySwapSwapIterator, error) {

	var fromTokenRule []interface{}
	for _, fromTokenItem := range fromToken {
		fromTokenRule = append(fromTokenRule, fromTokenItem)
	}
	var fromAmountRule []interface{}
	for _, fromAmountItem := range fromAmount {
		fromAmountRule = append(fromAmountRule, fromAmountItem)
	}

	logs, sub, err := _IcySwap.contract.FilterLogs(opts, "Swap", fromTokenRule, fromAmountRule)
	if err != nil {
		return nil, err
	}
	return &IcySwapSwapIterator{contract: _IcySwap.contract, event: "Swap", logs: logs, sub: sub}, nil
}

// WatchSwap is a free log subscription operation binding the contract event 0x562c219552544ec4c9d7a8eb850f80ea152973e315372bf4999fe7c953ea004f.
//
// Solidity: event Swap(address indexed fromToken, uint256 indexed fromAmount)
func (_IcySwap *IcySwapFilterer) WatchSwap(opts *bind.WatchOpts, sink chan<- *IcySwapSwap, fromToken []common.Address, fromAmount []*big.Int) (event.Subscription, error) {

	var fromTokenRule []interface{}
	for _, fromTokenItem := range fromToken {
		fromTokenRule = append(fromTokenRule, fromTokenItem)
	}
	var fromAmountRule []interface{}
	for _, fromAmountItem := range fromAmount {
		fromAmountRule = append(fromAmountRule, fromAmountItem)
	}

	logs, sub, err := _IcySwap.contract.WatchLogs(opts, "Swap", fromTokenRule, fromAmountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IcySwapSwap)
				if err := _IcySwap.contract.UnpackLog(event, "Swap", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSwap is a log parse operation binding the contract event 0x562c219552544ec4c9d7a8eb850f80ea152973e315372bf4999fe7c953ea004f.
//
// Solidity: event Swap(address indexed fromToken, uint256 indexed fromAmount)
func (_IcySwap *IcySwapFilterer) ParseSwap(log types.Log) (*IcySwapSwap, error) {
	event := new(IcySwapSwap)
	if err := _IcySwap.contract.UnpackLog(event, "Swap", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IcySwapUnpausedIterator is returned from FilterUnpaused and is used to iterate over the raw logs and unpacked data for Unpaused events raised by the IcySwap contract.
type IcySwapUnpausedIterator struct {
	Event *IcySwapUnpaused // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *IcySwapUnpausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IcySwapUnpaused)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(IcySwapUnpaused)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *IcySwapUnpausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IcySwapUnpausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IcySwapUnpaused represents a Unpaused event raised by the IcySwap contract.
type IcySwapUnpaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterUnpaused is a free log retrieval operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_IcySwap *IcySwapFilterer) FilterUnpaused(opts *bind.FilterOpts) (*IcySwapUnpausedIterator, error) {

	logs, sub, err := _IcySwap.contract.FilterLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return &IcySwapUnpausedIterator{contract: _IcySwap.contract, event: "Unpaused", logs: logs, sub: sub}, nil
}

// WatchUnpaused is a free log subscription operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_IcySwap *IcySwapFilterer) WatchUnpaused(opts *bind.WatchOpts, sink chan<- *IcySwapUnpaused) (event.Subscription, error) {

	logs, sub, err := _IcySwap.contract.WatchLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IcySwapUnpaused)
				if err := _IcySwap.contract.UnpackLog(event, "Unpaused", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUnpaused is a log parse operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_IcySwap *IcySwapFilterer) ParseUnpaused(log types.Log) (*IcySwapUnpaused, error) {
	event := new(IcySwapUnpaused)
	if err := _IcySwap.contract.UnpackLog(event, "Unpaused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IcySwapWithdrawToOwnerIterator is returned from FilterWithdrawToOwner and is used to iterate over the raw logs and unpacked data for WithdrawToOwner events raised by the IcySwap contract.
type IcySwapWithdrawToOwnerIterator struct {
	Event *IcySwapWithdrawToOwner // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *IcySwapWithdrawToOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IcySwapWithdrawToOwner)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(IcySwapWithdrawToOwner)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *IcySwapWithdrawToOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IcySwapWithdrawToOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IcySwapWithdrawToOwner represents a WithdrawToOwner event raised by the IcySwap contract.
type IcySwapWithdrawToOwner struct {
	Token  common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterWithdrawToOwner is a free log retrieval operation binding the contract event 0x5324e5ca3eab399efb9cff88b357827404aac06c9bebbd13d81f095576581988.
//
// Solidity: event WithdrawToOwner(address indexed token, uint256 amount)
func (_IcySwap *IcySwapFilterer) FilterWithdrawToOwner(opts *bind.FilterOpts, token []common.Address) (*IcySwapWithdrawToOwnerIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _IcySwap.contract.FilterLogs(opts, "WithdrawToOwner", tokenRule)
	if err != nil {
		return nil, err
	}
	return &IcySwapWithdrawToOwnerIterator{contract: _IcySwap.contract, event: "WithdrawToOwner", logs: logs, sub: sub}, nil
}

// WatchWithdrawToOwner is a free log subscription operation binding the contract event 0x5324e5ca3eab399efb9cff88b357827404aac06c9bebbd13d81f095576581988.
//
// Solidity: event WithdrawToOwner(address indexed token, uint256 amount)
func (_IcySwap *IcySwapFilterer) WatchWithdrawToOwner(opts *bind.WatchOpts, sink chan<- *IcySwapWithdrawToOwner, token []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _IcySwap.contract.WatchLogs(opts, "WithdrawToOwner", tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IcySwapWithdrawToOwner)
				if err := _IcySwap.contract.UnpackLog(event, "WithdrawToOwner", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWithdrawToOwner is a log parse operation binding the contract event 0x5324e5ca3eab399efb9cff88b357827404aac06c9bebbd13d81f095576581988.
//
// Solidity: event WithdrawToOwner(address indexed token, uint256 amount)
func (_IcySwap *IcySwapFilterer) ParseWithdrawToOwner(log types.Log) (*IcySwapWithdrawToOwner, error) {
	event := new(IcySwapWithdrawToOwner)
	if err := _IcySwap.contract.UnpackLog(event, "WithdrawToOwner", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
