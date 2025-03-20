// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package icyswapbtc

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

// IcyswapbtcMetaData contains all meta data concerning the Icyswapbtc contract.
var IcyswapbtcMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_icy\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"icyAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"btcAddress\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"btcAmount\",\"type\":\"uint256\"}],\"name\":\"RevertIcy\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"signerAddress\",\"type\":\"address\"}],\"name\":\"SetSigner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"icyAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"btcAddress\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"btcAmount\",\"type\":\"uint256\"}],\"name\":\"Swap\",\"type\":\"event\"},{\"stateMutability\":\"nonpayable\",\"type\":\"fallback\"},{\"inputs\":[],\"name\":\"REVERT_ICY_HASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SWAP_HASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"eip712Domain\",\"outputs\":[{\"internalType\":\"bytes1\",\"name\":\"fields\",\"type\":\"bytes1\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"version\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"verifyingContract\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"salt\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"extensions\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"icyAmount\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"btcAddress\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"btcAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"}],\"name\":\"getRevertIcyHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_digest\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"_signature\",\"type\":\"bytes\"}],\"name\":\"getSigner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"icyAmount\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"btcAddress\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"btcAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"}],\"name\":\"getSwapHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"icy\",\"outputs\":[{\"internalType\":\"contractERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"icyAmount\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"btcAddress\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"btcAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_signature\",\"type\":\"bytes\"}],\"name\":\"revertIcy\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"revertedIcyHashes\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_signerAddress\",\"type\":\"address\"}],\"name\":\"setSigner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"signerAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"icyAmount\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"btcAddress\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"btcAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_signature\",\"type\":\"bytes\"}],\"name\":\"swap\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"swappedHashes\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
}

// IcyswapbtcABI is the input ABI used to generate the binding from.
// Deprecated: Use IcyswapbtcMetaData.ABI instead.
var IcyswapbtcABI = IcyswapbtcMetaData.ABI

// Icyswapbtc is an auto generated Go binding around an Ethereum contract.
type Icyswapbtc struct {
	IcyswapbtcCaller     // Read-only binding to the contract
	IcyswapbtcTransactor // Write-only binding to the contract
	IcyswapbtcFilterer   // Log filterer for contract events
}

// IcyswapbtcCaller is an auto generated read-only Go binding around an Ethereum contract.
type IcyswapbtcCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IcyswapbtcTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IcyswapbtcTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IcyswapbtcFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IcyswapbtcFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IcyswapbtcSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IcyswapbtcSession struct {
	Contract     *Icyswapbtc       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IcyswapbtcCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IcyswapbtcCallerSession struct {
	Contract *IcyswapbtcCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// IcyswapbtcTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IcyswapbtcTransactorSession struct {
	Contract     *IcyswapbtcTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// IcyswapbtcRaw is an auto generated low-level Go binding around an Ethereum contract.
type IcyswapbtcRaw struct {
	Contract *Icyswapbtc // Generic contract binding to access the raw methods on
}

// IcyswapbtcCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IcyswapbtcCallerRaw struct {
	Contract *IcyswapbtcCaller // Generic read-only contract binding to access the raw methods on
}

// IcyswapbtcTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IcyswapbtcTransactorRaw struct {
	Contract *IcyswapbtcTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIcyswapbtc creates a new instance of Icyswapbtc, bound to a specific deployed contract.
func NewIcyswapbtc(address common.Address, backend bind.ContractBackend) (*Icyswapbtc, error) {
	contract, err := bindIcyswapbtc(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Icyswapbtc{IcyswapbtcCaller: IcyswapbtcCaller{contract: contract}, IcyswapbtcTransactor: IcyswapbtcTransactor{contract: contract}, IcyswapbtcFilterer: IcyswapbtcFilterer{contract: contract}}, nil
}

// NewIcyswapbtcCaller creates a new read-only instance of Icyswapbtc, bound to a specific deployed contract.
func NewIcyswapbtcCaller(address common.Address, caller bind.ContractCaller) (*IcyswapbtcCaller, error) {
	contract, err := bindIcyswapbtc(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IcyswapbtcCaller{contract: contract}, nil
}

// NewIcyswapbtcTransactor creates a new write-only instance of Icyswapbtc, bound to a specific deployed contract.
func NewIcyswapbtcTransactor(address common.Address, transactor bind.ContractTransactor) (*IcyswapbtcTransactor, error) {
	contract, err := bindIcyswapbtc(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IcyswapbtcTransactor{contract: contract}, nil
}

// NewIcyswapbtcFilterer creates a new log filterer instance of Icyswapbtc, bound to a specific deployed contract.
func NewIcyswapbtcFilterer(address common.Address, filterer bind.ContractFilterer) (*IcyswapbtcFilterer, error) {
	contract, err := bindIcyswapbtc(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IcyswapbtcFilterer{contract: contract}, nil
}

// bindIcyswapbtc binds a generic wrapper to an already deployed contract.
func bindIcyswapbtc(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IcyswapbtcABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Icyswapbtc *IcyswapbtcRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Icyswapbtc.Contract.IcyswapbtcCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Icyswapbtc *IcyswapbtcRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Icyswapbtc.Contract.IcyswapbtcTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Icyswapbtc *IcyswapbtcRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Icyswapbtc.Contract.IcyswapbtcTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Icyswapbtc *IcyswapbtcCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Icyswapbtc.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Icyswapbtc *IcyswapbtcTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Icyswapbtc.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Icyswapbtc *IcyswapbtcTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Icyswapbtc.Contract.contract.Transact(opts, method, params...)
}

// REVERTICYHASH is a free data retrieval call binding the contract method 0x2d6d3d01.
//
// Solidity: function REVERT_ICY_HASH() view returns(bytes32)
func (_Icyswapbtc *IcyswapbtcCaller) REVERTICYHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Icyswapbtc.contract.Call(opts, &out, "REVERT_ICY_HASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// REVERTICYHASH is a free data retrieval call binding the contract method 0x2d6d3d01.
//
// Solidity: function REVERT_ICY_HASH() view returns(bytes32)
func (_Icyswapbtc *IcyswapbtcSession) REVERTICYHASH() ([32]byte, error) {
	return _Icyswapbtc.Contract.REVERTICYHASH(&_Icyswapbtc.CallOpts)
}

// REVERTICYHASH is a free data retrieval call binding the contract method 0x2d6d3d01.
//
// Solidity: function REVERT_ICY_HASH() view returns(bytes32)
func (_Icyswapbtc *IcyswapbtcCallerSession) REVERTICYHASH() ([32]byte, error) {
	return _Icyswapbtc.Contract.REVERTICYHASH(&_Icyswapbtc.CallOpts)
}

// SWAPHASH is a free data retrieval call binding the contract method 0x30c8b3da.
//
// Solidity: function SWAP_HASH() view returns(bytes32)
func (_Icyswapbtc *IcyswapbtcCaller) SWAPHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Icyswapbtc.contract.Call(opts, &out, "SWAP_HASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SWAPHASH is a free data retrieval call binding the contract method 0x30c8b3da.
//
// Solidity: function SWAP_HASH() view returns(bytes32)
func (_Icyswapbtc *IcyswapbtcSession) SWAPHASH() ([32]byte, error) {
	return _Icyswapbtc.Contract.SWAPHASH(&_Icyswapbtc.CallOpts)
}

// SWAPHASH is a free data retrieval call binding the contract method 0x30c8b3da.
//
// Solidity: function SWAP_HASH() view returns(bytes32)
func (_Icyswapbtc *IcyswapbtcCallerSession) SWAPHASH() ([32]byte, error) {
	return _Icyswapbtc.Contract.SWAPHASH(&_Icyswapbtc.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_Icyswapbtc *IcyswapbtcCaller) Eip712Domain(opts *bind.CallOpts) (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	var out []interface{}
	err := _Icyswapbtc.contract.Call(opts, &out, "eip712Domain")

	outstruct := new(struct {
		Fields            [1]byte
		Name              string
		Version           string
		ChainId           *big.Int
		VerifyingContract common.Address
		Salt              [32]byte
		Extensions        []*big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Fields = *abi.ConvertType(out[0], new([1]byte)).(*[1]byte)
	outstruct.Name = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.Version = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.ChainId = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.VerifyingContract = *abi.ConvertType(out[4], new(common.Address)).(*common.Address)
	outstruct.Salt = *abi.ConvertType(out[5], new([32]byte)).(*[32]byte)
	outstruct.Extensions = *abi.ConvertType(out[6], new([]*big.Int)).(*[]*big.Int)

	return *outstruct, err

}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_Icyswapbtc *IcyswapbtcSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _Icyswapbtc.Contract.Eip712Domain(&_Icyswapbtc.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_Icyswapbtc *IcyswapbtcCallerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _Icyswapbtc.Contract.Eip712Domain(&_Icyswapbtc.CallOpts)
}

// GetRevertIcyHash is a free data retrieval call binding the contract method 0x32ce558d.
//
// Solidity: function getRevertIcyHash(uint256 icyAmount, string btcAddress, uint256 btcAmount, uint256 nonce, uint256 deadline) view returns(bytes32 hash)
func (_Icyswapbtc *IcyswapbtcCaller) GetRevertIcyHash(opts *bind.CallOpts, icyAmount *big.Int, btcAddress string, btcAmount *big.Int, nonce *big.Int, deadline *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _Icyswapbtc.contract.Call(opts, &out, "getRevertIcyHash", icyAmount, btcAddress, btcAmount, nonce, deadline)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRevertIcyHash is a free data retrieval call binding the contract method 0x32ce558d.
//
// Solidity: function getRevertIcyHash(uint256 icyAmount, string btcAddress, uint256 btcAmount, uint256 nonce, uint256 deadline) view returns(bytes32 hash)
func (_Icyswapbtc *IcyswapbtcSession) GetRevertIcyHash(icyAmount *big.Int, btcAddress string, btcAmount *big.Int, nonce *big.Int, deadline *big.Int) ([32]byte, error) {
	return _Icyswapbtc.Contract.GetRevertIcyHash(&_Icyswapbtc.CallOpts, icyAmount, btcAddress, btcAmount, nonce, deadline)
}

// GetRevertIcyHash is a free data retrieval call binding the contract method 0x32ce558d.
//
// Solidity: function getRevertIcyHash(uint256 icyAmount, string btcAddress, uint256 btcAmount, uint256 nonce, uint256 deadline) view returns(bytes32 hash)
func (_Icyswapbtc *IcyswapbtcCallerSession) GetRevertIcyHash(icyAmount *big.Int, btcAddress string, btcAmount *big.Int, nonce *big.Int, deadline *big.Int) ([32]byte, error) {
	return _Icyswapbtc.Contract.GetRevertIcyHash(&_Icyswapbtc.CallOpts, icyAmount, btcAddress, btcAmount, nonce, deadline)
}

// GetSigner is a free data retrieval call binding the contract method 0xf7b2ec0d.
//
// Solidity: function getSigner(bytes32 _digest, bytes _signature) view returns(address)
func (_Icyswapbtc *IcyswapbtcCaller) GetSigner(opts *bind.CallOpts, _digest [32]byte, _signature []byte) (common.Address, error) {
	var out []interface{}
	err := _Icyswapbtc.contract.Call(opts, &out, "getSigner", _digest, _signature)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetSigner is a free data retrieval call binding the contract method 0xf7b2ec0d.
//
// Solidity: function getSigner(bytes32 _digest, bytes _signature) view returns(address)
func (_Icyswapbtc *IcyswapbtcSession) GetSigner(_digest [32]byte, _signature []byte) (common.Address, error) {
	return _Icyswapbtc.Contract.GetSigner(&_Icyswapbtc.CallOpts, _digest, _signature)
}

// GetSigner is a free data retrieval call binding the contract method 0xf7b2ec0d.
//
// Solidity: function getSigner(bytes32 _digest, bytes _signature) view returns(address)
func (_Icyswapbtc *IcyswapbtcCallerSession) GetSigner(_digest [32]byte, _signature []byte) (common.Address, error) {
	return _Icyswapbtc.Contract.GetSigner(&_Icyswapbtc.CallOpts, _digest, _signature)
}

// GetSwapHash is a free data retrieval call binding the contract method 0x6327a9d0.
//
// Solidity: function getSwapHash(uint256 icyAmount, string btcAddress, uint256 btcAmount, uint256 nonce, uint256 deadline) view returns(bytes32 hash)
func (_Icyswapbtc *IcyswapbtcCaller) GetSwapHash(opts *bind.CallOpts, icyAmount *big.Int, btcAddress string, btcAmount *big.Int, nonce *big.Int, deadline *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _Icyswapbtc.contract.Call(opts, &out, "getSwapHash", icyAmount, btcAddress, btcAmount, nonce, deadline)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetSwapHash is a free data retrieval call binding the contract method 0x6327a9d0.
//
// Solidity: function getSwapHash(uint256 icyAmount, string btcAddress, uint256 btcAmount, uint256 nonce, uint256 deadline) view returns(bytes32 hash)
func (_Icyswapbtc *IcyswapbtcSession) GetSwapHash(icyAmount *big.Int, btcAddress string, btcAmount *big.Int, nonce *big.Int, deadline *big.Int) ([32]byte, error) {
	return _Icyswapbtc.Contract.GetSwapHash(&_Icyswapbtc.CallOpts, icyAmount, btcAddress, btcAmount, nonce, deadline)
}

// GetSwapHash is a free data retrieval call binding the contract method 0x6327a9d0.
//
// Solidity: function getSwapHash(uint256 icyAmount, string btcAddress, uint256 btcAmount, uint256 nonce, uint256 deadline) view returns(bytes32 hash)
func (_Icyswapbtc *IcyswapbtcCallerSession) GetSwapHash(icyAmount *big.Int, btcAddress string, btcAmount *big.Int, nonce *big.Int, deadline *big.Int) ([32]byte, error) {
	return _Icyswapbtc.Contract.GetSwapHash(&_Icyswapbtc.CallOpts, icyAmount, btcAddress, btcAmount, nonce, deadline)
}

// Icy is a free data retrieval call binding the contract method 0x7f245ab1.
//
// Solidity: function icy() view returns(address)
func (_Icyswapbtc *IcyswapbtcCaller) Icy(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Icyswapbtc.contract.Call(opts, &out, "icy")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Icy is a free data retrieval call binding the contract method 0x7f245ab1.
//
// Solidity: function icy() view returns(address)
func (_Icyswapbtc *IcyswapbtcSession) Icy() (common.Address, error) {
	return _Icyswapbtc.Contract.Icy(&_Icyswapbtc.CallOpts)
}

// Icy is a free data retrieval call binding the contract method 0x7f245ab1.
//
// Solidity: function icy() view returns(address)
func (_Icyswapbtc *IcyswapbtcCallerSession) Icy() (common.Address, error) {
	return _Icyswapbtc.Contract.Icy(&_Icyswapbtc.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Icyswapbtc *IcyswapbtcCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Icyswapbtc.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Icyswapbtc *IcyswapbtcSession) Owner() (common.Address, error) {
	return _Icyswapbtc.Contract.Owner(&_Icyswapbtc.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Icyswapbtc *IcyswapbtcCallerSession) Owner() (common.Address, error) {
	return _Icyswapbtc.Contract.Owner(&_Icyswapbtc.CallOpts)
}

// RevertedIcyHashes is a free data retrieval call binding the contract method 0x3d2b52db.
//
// Solidity: function revertedIcyHashes(bytes32 ) view returns(bool)
func (_Icyswapbtc *IcyswapbtcCaller) RevertedIcyHashes(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var out []interface{}
	err := _Icyswapbtc.contract.Call(opts, &out, "revertedIcyHashes", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// RevertedIcyHashes is a free data retrieval call binding the contract method 0x3d2b52db.
//
// Solidity: function revertedIcyHashes(bytes32 ) view returns(bool)
func (_Icyswapbtc *IcyswapbtcSession) RevertedIcyHashes(arg0 [32]byte) (bool, error) {
	return _Icyswapbtc.Contract.RevertedIcyHashes(&_Icyswapbtc.CallOpts, arg0)
}

// RevertedIcyHashes is a free data retrieval call binding the contract method 0x3d2b52db.
//
// Solidity: function revertedIcyHashes(bytes32 ) view returns(bool)
func (_Icyswapbtc *IcyswapbtcCallerSession) RevertedIcyHashes(arg0 [32]byte) (bool, error) {
	return _Icyswapbtc.Contract.RevertedIcyHashes(&_Icyswapbtc.CallOpts, arg0)
}

// SignerAddress is a free data retrieval call binding the contract method 0x5b7633d0.
//
// Solidity: function signerAddress() view returns(address)
func (_Icyswapbtc *IcyswapbtcCaller) SignerAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Icyswapbtc.contract.Call(opts, &out, "signerAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// SignerAddress is a free data retrieval call binding the contract method 0x5b7633d0.
//
// Solidity: function signerAddress() view returns(address)
func (_Icyswapbtc *IcyswapbtcSession) SignerAddress() (common.Address, error) {
	return _Icyswapbtc.Contract.SignerAddress(&_Icyswapbtc.CallOpts)
}

// SignerAddress is a free data retrieval call binding the contract method 0x5b7633d0.
//
// Solidity: function signerAddress() view returns(address)
func (_Icyswapbtc *IcyswapbtcCallerSession) SignerAddress() (common.Address, error) {
	return _Icyswapbtc.Contract.SignerAddress(&_Icyswapbtc.CallOpts)
}

// SwappedHashes is a free data retrieval call binding the contract method 0x6072e236.
//
// Solidity: function swappedHashes(bytes32 ) view returns(bool)
func (_Icyswapbtc *IcyswapbtcCaller) SwappedHashes(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var out []interface{}
	err := _Icyswapbtc.contract.Call(opts, &out, "swappedHashes", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SwappedHashes is a free data retrieval call binding the contract method 0x6072e236.
//
// Solidity: function swappedHashes(bytes32 ) view returns(bool)
func (_Icyswapbtc *IcyswapbtcSession) SwappedHashes(arg0 [32]byte) (bool, error) {
	return _Icyswapbtc.Contract.SwappedHashes(&_Icyswapbtc.CallOpts, arg0)
}

// SwappedHashes is a free data retrieval call binding the contract method 0x6072e236.
//
// Solidity: function swappedHashes(bytes32 ) view returns(bool)
func (_Icyswapbtc *IcyswapbtcCallerSession) SwappedHashes(arg0 [32]byte) (bool, error) {
	return _Icyswapbtc.Contract.SwappedHashes(&_Icyswapbtc.CallOpts, arg0)
}

// RevertIcy is a paid mutator transaction binding the contract method 0x666f5a65.
//
// Solidity: function revertIcy(uint256 icyAmount, string btcAddress, uint256 btcAmount, uint256 nonce, uint256 deadline, bytes _signature) returns()
func (_Icyswapbtc *IcyswapbtcTransactor) RevertIcy(opts *bind.TransactOpts, icyAmount *big.Int, btcAddress string, btcAmount *big.Int, nonce *big.Int, deadline *big.Int, _signature []byte) (*types.Transaction, error) {
	return _Icyswapbtc.contract.Transact(opts, "revertIcy", icyAmount, btcAddress, btcAmount, nonce, deadline, _signature)
}

// RevertIcy is a paid mutator transaction binding the contract method 0x666f5a65.
//
// Solidity: function revertIcy(uint256 icyAmount, string btcAddress, uint256 btcAmount, uint256 nonce, uint256 deadline, bytes _signature) returns()
func (_Icyswapbtc *IcyswapbtcSession) RevertIcy(icyAmount *big.Int, btcAddress string, btcAmount *big.Int, nonce *big.Int, deadline *big.Int, _signature []byte) (*types.Transaction, error) {
	return _Icyswapbtc.Contract.RevertIcy(&_Icyswapbtc.TransactOpts, icyAmount, btcAddress, btcAmount, nonce, deadline, _signature)
}

// RevertIcy is a paid mutator transaction binding the contract method 0x666f5a65.
//
// Solidity: function revertIcy(uint256 icyAmount, string btcAddress, uint256 btcAmount, uint256 nonce, uint256 deadline, bytes _signature) returns()
func (_Icyswapbtc *IcyswapbtcTransactorSession) RevertIcy(icyAmount *big.Int, btcAddress string, btcAmount *big.Int, nonce *big.Int, deadline *big.Int, _signature []byte) (*types.Transaction, error) {
	return _Icyswapbtc.Contract.RevertIcy(&_Icyswapbtc.TransactOpts, icyAmount, btcAddress, btcAmount, nonce, deadline, _signature)
}

// SetSigner is a paid mutator transaction binding the contract method 0x6c19e783.
//
// Solidity: function setSigner(address _signerAddress) returns()
func (_Icyswapbtc *IcyswapbtcTransactor) SetSigner(opts *bind.TransactOpts, _signerAddress common.Address) (*types.Transaction, error) {
	return _Icyswapbtc.contract.Transact(opts, "setSigner", _signerAddress)
}

// SetSigner is a paid mutator transaction binding the contract method 0x6c19e783.
//
// Solidity: function setSigner(address _signerAddress) returns()
func (_Icyswapbtc *IcyswapbtcSession) SetSigner(_signerAddress common.Address) (*types.Transaction, error) {
	return _Icyswapbtc.Contract.SetSigner(&_Icyswapbtc.TransactOpts, _signerAddress)
}

// SetSigner is a paid mutator transaction binding the contract method 0x6c19e783.
//
// Solidity: function setSigner(address _signerAddress) returns()
func (_Icyswapbtc *IcyswapbtcTransactorSession) SetSigner(_signerAddress common.Address) (*types.Transaction, error) {
	return _Icyswapbtc.Contract.SetSigner(&_Icyswapbtc.TransactOpts, _signerAddress)
}

// Swap is a paid mutator transaction binding the contract method 0xade44138.
//
// Solidity: function swap(uint256 icyAmount, string btcAddress, uint256 btcAmount, uint256 nonce, uint256 deadline, bytes _signature) returns()
func (_Icyswapbtc *IcyswapbtcTransactor) Swap(opts *bind.TransactOpts, icyAmount *big.Int, btcAddress string, btcAmount *big.Int, nonce *big.Int, deadline *big.Int, _signature []byte) (*types.Transaction, error) {
	return _Icyswapbtc.contract.Transact(opts, "swap", icyAmount, btcAddress, btcAmount, nonce, deadline, _signature)
}

// Swap is a paid mutator transaction binding the contract method 0xade44138.
//
// Solidity: function swap(uint256 icyAmount, string btcAddress, uint256 btcAmount, uint256 nonce, uint256 deadline, bytes _signature) returns()
func (_Icyswapbtc *IcyswapbtcSession) Swap(icyAmount *big.Int, btcAddress string, btcAmount *big.Int, nonce *big.Int, deadline *big.Int, _signature []byte) (*types.Transaction, error) {
	return _Icyswapbtc.Contract.Swap(&_Icyswapbtc.TransactOpts, icyAmount, btcAddress, btcAmount, nonce, deadline, _signature)
}

// Swap is a paid mutator transaction binding the contract method 0xade44138.
//
// Solidity: function swap(uint256 icyAmount, string btcAddress, uint256 btcAmount, uint256 nonce, uint256 deadline, bytes _signature) returns()
func (_Icyswapbtc *IcyswapbtcTransactorSession) Swap(icyAmount *big.Int, btcAddress string, btcAmount *big.Int, nonce *big.Int, deadline *big.Int, _signature []byte) (*types.Transaction, error) {
	return _Icyswapbtc.Contract.Swap(&_Icyswapbtc.TransactOpts, icyAmount, btcAddress, btcAmount, nonce, deadline, _signature)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Icyswapbtc *IcyswapbtcTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Icyswapbtc.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Icyswapbtc *IcyswapbtcSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Icyswapbtc.Contract.TransferOwnership(&_Icyswapbtc.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Icyswapbtc *IcyswapbtcTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Icyswapbtc.Contract.TransferOwnership(&_Icyswapbtc.TransactOpts, newOwner)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_Icyswapbtc *IcyswapbtcTransactor) Fallback(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error) {
	return _Icyswapbtc.contract.RawTransact(opts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_Icyswapbtc *IcyswapbtcSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Icyswapbtc.Contract.Fallback(&_Icyswapbtc.TransactOpts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_Icyswapbtc *IcyswapbtcTransactorSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Icyswapbtc.Contract.Fallback(&_Icyswapbtc.TransactOpts, calldata)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Icyswapbtc *IcyswapbtcTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Icyswapbtc.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Icyswapbtc *IcyswapbtcSession) Receive() (*types.Transaction, error) {
	return _Icyswapbtc.Contract.Receive(&_Icyswapbtc.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Icyswapbtc *IcyswapbtcTransactorSession) Receive() (*types.Transaction, error) {
	return _Icyswapbtc.Contract.Receive(&_Icyswapbtc.TransactOpts)
}

// IcyswapbtcOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Icyswapbtc contract.
type IcyswapbtcOwnershipTransferredIterator struct {
	Event *IcyswapbtcOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *IcyswapbtcOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IcyswapbtcOwnershipTransferred)
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
		it.Event = new(IcyswapbtcOwnershipTransferred)
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
func (it *IcyswapbtcOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IcyswapbtcOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IcyswapbtcOwnershipTransferred represents a OwnershipTransferred event raised by the Icyswapbtc contract.
type IcyswapbtcOwnershipTransferred struct {
	User     common.Address
	NewOwner common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed user, address indexed newOwner)
func (_Icyswapbtc *IcyswapbtcFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, user []common.Address, newOwner []common.Address) (*IcyswapbtcOwnershipTransferredIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Icyswapbtc.contract.FilterLogs(opts, "OwnershipTransferred", userRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &IcyswapbtcOwnershipTransferredIterator{contract: _Icyswapbtc.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed user, address indexed newOwner)
func (_Icyswapbtc *IcyswapbtcFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *IcyswapbtcOwnershipTransferred, user []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Icyswapbtc.contract.WatchLogs(opts, "OwnershipTransferred", userRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IcyswapbtcOwnershipTransferred)
				if err := _Icyswapbtc.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
// Solidity: event OwnershipTransferred(address indexed user, address indexed newOwner)
func (_Icyswapbtc *IcyswapbtcFilterer) ParseOwnershipTransferred(log types.Log) (*IcyswapbtcOwnershipTransferred, error) {
	event := new(IcyswapbtcOwnershipTransferred)
	if err := _Icyswapbtc.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IcyswapbtcRevertIcyIterator is returned from FilterRevertIcy and is used to iterate over the raw logs and unpacked data for RevertIcy events raised by the Icyswapbtc contract.
type IcyswapbtcRevertIcyIterator struct {
	Event *IcyswapbtcRevertIcy // Event containing the contract specifics and raw log

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
func (it *IcyswapbtcRevertIcyIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IcyswapbtcRevertIcy)
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
		it.Event = new(IcyswapbtcRevertIcy)
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
func (it *IcyswapbtcRevertIcyIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IcyswapbtcRevertIcyIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IcyswapbtcRevertIcy represents a RevertIcy event raised by the Icyswapbtc contract.
type IcyswapbtcRevertIcy struct {
	IcyAmount  *big.Int
	BtcAddress string
	BtcAmount  *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterRevertIcy is a free log retrieval operation binding the contract event 0xd65289b780c2a5756f2385450f37835d3af0fd779700af98d868c8f952e9acff.
//
// Solidity: event RevertIcy(uint256 icyAmount, string btcAddress, uint256 btcAmount)
func (_Icyswapbtc *IcyswapbtcFilterer) FilterRevertIcy(opts *bind.FilterOpts) (*IcyswapbtcRevertIcyIterator, error) {

	logs, sub, err := _Icyswapbtc.contract.FilterLogs(opts, "RevertIcy")
	if err != nil {
		return nil, err
	}
	return &IcyswapbtcRevertIcyIterator{contract: _Icyswapbtc.contract, event: "RevertIcy", logs: logs, sub: sub}, nil
}

// WatchRevertIcy is a free log subscription operation binding the contract event 0xd65289b780c2a5756f2385450f37835d3af0fd779700af98d868c8f952e9acff.
//
// Solidity: event RevertIcy(uint256 icyAmount, string btcAddress, uint256 btcAmount)
func (_Icyswapbtc *IcyswapbtcFilterer) WatchRevertIcy(opts *bind.WatchOpts, sink chan<- *IcyswapbtcRevertIcy) (event.Subscription, error) {

	logs, sub, err := _Icyswapbtc.contract.WatchLogs(opts, "RevertIcy")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IcyswapbtcRevertIcy)
				if err := _Icyswapbtc.contract.UnpackLog(event, "RevertIcy", log); err != nil {
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

// ParseRevertIcy is a log parse operation binding the contract event 0xd65289b780c2a5756f2385450f37835d3af0fd779700af98d868c8f952e9acff.
//
// Solidity: event RevertIcy(uint256 icyAmount, string btcAddress, uint256 btcAmount)
func (_Icyswapbtc *IcyswapbtcFilterer) ParseRevertIcy(log types.Log) (*IcyswapbtcRevertIcy, error) {
	event := new(IcyswapbtcRevertIcy)
	if err := _Icyswapbtc.contract.UnpackLog(event, "RevertIcy", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IcyswapbtcSetSignerIterator is returned from FilterSetSigner and is used to iterate over the raw logs and unpacked data for SetSigner events raised by the Icyswapbtc contract.
type IcyswapbtcSetSignerIterator struct {
	Event *IcyswapbtcSetSigner // Event containing the contract specifics and raw log

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
func (it *IcyswapbtcSetSignerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IcyswapbtcSetSigner)
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
		it.Event = new(IcyswapbtcSetSigner)
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
func (it *IcyswapbtcSetSignerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IcyswapbtcSetSignerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IcyswapbtcSetSigner represents a SetSigner event raised by the Icyswapbtc contract.
type IcyswapbtcSetSigner struct {
	SignerAddress common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterSetSigner is a free log retrieval operation binding the contract event 0xbb10aee7ef5a307b8097c6a7f2892b909ff1736fd24a6a5260640c185f7153b6.
//
// Solidity: event SetSigner(address signerAddress)
func (_Icyswapbtc *IcyswapbtcFilterer) FilterSetSigner(opts *bind.FilterOpts) (*IcyswapbtcSetSignerIterator, error) {

	logs, sub, err := _Icyswapbtc.contract.FilterLogs(opts, "SetSigner")
	if err != nil {
		return nil, err
	}
	return &IcyswapbtcSetSignerIterator{contract: _Icyswapbtc.contract, event: "SetSigner", logs: logs, sub: sub}, nil
}

// WatchSetSigner is a free log subscription operation binding the contract event 0xbb10aee7ef5a307b8097c6a7f2892b909ff1736fd24a6a5260640c185f7153b6.
//
// Solidity: event SetSigner(address signerAddress)
func (_Icyswapbtc *IcyswapbtcFilterer) WatchSetSigner(opts *bind.WatchOpts, sink chan<- *IcyswapbtcSetSigner) (event.Subscription, error) {

	logs, sub, err := _Icyswapbtc.contract.WatchLogs(opts, "SetSigner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IcyswapbtcSetSigner)
				if err := _Icyswapbtc.contract.UnpackLog(event, "SetSigner", log); err != nil {
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

// ParseSetSigner is a log parse operation binding the contract event 0xbb10aee7ef5a307b8097c6a7f2892b909ff1736fd24a6a5260640c185f7153b6.
//
// Solidity: event SetSigner(address signerAddress)
func (_Icyswapbtc *IcyswapbtcFilterer) ParseSetSigner(log types.Log) (*IcyswapbtcSetSigner, error) {
	event := new(IcyswapbtcSetSigner)
	if err := _Icyswapbtc.contract.UnpackLog(event, "SetSigner", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IcyswapbtcSwapIterator is returned from FilterSwap and is used to iterate over the raw logs and unpacked data for Swap events raised by the Icyswapbtc contract.
type IcyswapbtcSwapIterator struct {
	Event *IcyswapbtcSwap // Event containing the contract specifics and raw log

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
func (it *IcyswapbtcSwapIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IcyswapbtcSwap)
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
		it.Event = new(IcyswapbtcSwap)
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
func (it *IcyswapbtcSwapIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IcyswapbtcSwapIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IcyswapbtcSwap represents a Swap event raised by the Icyswapbtc contract.
type IcyswapbtcSwap struct {
	IcyAmount  *big.Int
	BtcAddress string
	BtcAmount  *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterSwap is a free log retrieval operation binding the contract event 0x6a7e3add5ba4ffd84c70888b34c2abc4eb346e94dbd82a1ba6dd8e335c682063.
//
// Solidity: event Swap(uint256 icyAmount, string btcAddress, uint256 btcAmount)
func (_Icyswapbtc *IcyswapbtcFilterer) FilterSwap(opts *bind.FilterOpts) (*IcyswapbtcSwapIterator, error) {

	logs, sub, err := _Icyswapbtc.contract.FilterLogs(opts, "Swap")
	if err != nil {
		return nil, err
	}
	return &IcyswapbtcSwapIterator{contract: _Icyswapbtc.contract, event: "Swap", logs: logs, sub: sub}, nil
}

// WatchSwap is a free log subscription operation binding the contract event 0x6a7e3add5ba4ffd84c70888b34c2abc4eb346e94dbd82a1ba6dd8e335c682063.
//
// Solidity: event Swap(uint256 icyAmount, string btcAddress, uint256 btcAmount)
func (_Icyswapbtc *IcyswapbtcFilterer) WatchSwap(opts *bind.WatchOpts, sink chan<- *IcyswapbtcSwap) (event.Subscription, error) {

	logs, sub, err := _Icyswapbtc.contract.WatchLogs(opts, "Swap")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IcyswapbtcSwap)
				if err := _Icyswapbtc.contract.UnpackLog(event, "Swap", log); err != nil {
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

// ParseSwap is a log parse operation binding the contract event 0x6a7e3add5ba4ffd84c70888b34c2abc4eb346e94dbd82a1ba6dd8e335c682063.
//
// Solidity: event Swap(uint256 icyAmount, string btcAddress, uint256 btcAmount)
func (_Icyswapbtc *IcyswapbtcFilterer) ParseSwap(log types.Log) (*IcyswapbtcSwap, error) {
	event := new(IcyswapbtcSwap)
	if err := _Icyswapbtc.contract.UnpackLog(event, "Swap", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
