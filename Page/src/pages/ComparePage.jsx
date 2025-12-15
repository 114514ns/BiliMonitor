import React from 'react';
import {
    Select,
    SelectItem,
    Table,
    TableBody,
    TableCell,
    TableColumn,
    TableHeader,
    TableRow,
} from "@heroui/react";
import axios from "axios";
import { useAsyncList } from "@react-stately/data";
import {NavLink} from "react-router-dom";

function ComparePage() {
    const [mode, setMode] = React.useState("range")

    const [guild,setGuild] = React.useState("1")

    const [month,setMonth] = React.useState(new Date().getMonth() + 1 + '')
    const list = useAsyncList({
        async load({ signal }) {
            const res = await axios.get(`/api/pk?guild=${guild}&month=${month}`, { signal });
            res.data.data.sort((a,b) => calc(b)-calc(a))
            return { items: res.data.data };
        },
        async sort({ items, sortDescriptor }) {
            const sorted = [...items].sort((a, b) => {
                var first = a[sortDescriptor.column]
                var second = b[sortDescriptor.column]

                if (sortDescriptor.column === 'Money') {
                    first = calc(a)
                    second = calc(b)
                }

                let cmp = 0;
                if (!isNaN(parseFloat(first)) && !isNaN(parseFloat(second))) {
                    cmp = parseFloat(first) - parseFloat(second);
                } else {
                    cmp = String(first).localeCompare(String(second));
                }

                if (sortDescriptor.direction === "descending") {
                    cmp *= -1;
                }
                return cmp;
            });

            items.forEach(item => {
                item.Money = calc(item)
            })

            return { items: sorted };
        },
    });

    React.useEffect(() => {
        list.reload()
    },[guild,mode,month])

    return (
        <div>
            <div className="flex gap-4 items-center flex-col lg:flex-row">
                <Select
                    className="max-w-xs lg:ml-4 sm:mt-2"
                    label="Dimension"
                    selectedKeys={[mode]}
                    onSelectionChange={(keys) => setMode(Array.from(keys)[0])}
                >
                    <SelectItem key="range">By Range</SelectItem>
                    <SelectItem key="pk">PK</SelectItem>
                </Select>

                {mode === "range" && (
                    <Select
                        className="max-w-xs lg:ml-4 sm:mt-2"
                        label="Guild"
                        defaultSelectedKeys={['1']}
                        onSelectionChange={(keys) => {
                            const value = Array.from(keys)[0];
                            setGuild(value);
                        }}
                    >
                        <SelectItem key="1">VirtuaReal</SelectItem>
                        <SelectItem key="2">PSPLive</SelectItem>
                        <SelectItem key="">Any</SelectItem>
                    </Select>
                )}
                {mode === "range" && (
                    <Select
                        className="max-w-xs lg:ml-4 sm:mt-2"
                        label="Month"
                        defaultSelectedKeys={[month]}
                        selectedKeys={[month]}
                    >
                        {['1','2','3','4','5','6','7','8','9','10',11,'12'].map((item) => {
                            return <SelectItem key={item} onClick={(e) => {
                                setMonth(e.target.textContent)
                            }}>{item}</SelectItem>
                        })}
                    </Select>
                )}
            </div>

            {mode === "range" && (
                <div className="mt-2">
                    <Table
                        allowsSorting
                        isStriped
                        sortDescriptor={list.sortDescriptor}
                        onSortChange={(sort) => list.sort(sort)}
                    >
                        <TableHeader>
                            <TableColumn>Liver</TableColumn>
                            <TableColumn key="Money" allowsSorting>
                                总营收
                            </TableColumn>
                            <TableColumn key="SuperChatMoney" allowsSorting>
                                SuperChat营收
                            </TableColumn>
                            <TableColumn key="Gift" allowsSorting>
                                礼物营收
                            </TableColumn>
                            <TableColumn key="Guard" allowsSorting>
                                大航海
                            </TableColumn>
                            <TableColumn key="Hours" allowsSorting>
                                直播时长
                            </TableColumn>
                            <TableColumn key="Guild" allowsSorting>
                                公会
                            </TableColumn>
                        </TableHeader>
                        <TableBody
                            items={list.items}
                            isLoading={list.isLoading}
                            emptyContent={"No data"}
                        >
                            {(item) => (
                                <TableRow key={item.UserID}>
                                    <TableCell>
                                        <div className="flex flex-row items-center hover:text-gray-500">
                                            <img
                                                src={`${AVATAR_API}${item.UserID}`}
                                                className="h-[40px] w-[40px] rounded-full"
                                            />
                                            <NavLink className="ml-2" to={`/liver/${item.UserID}`}>{item.UserName}</NavLink>
                                        </div>
                                    </TableCell>
                                    <TableCell>{formatNumber(calc(item))}</TableCell>
                                    <TableCell>{formatNumber(item.SuperChatMoney)}</TableCell>
                                    <TableCell>{formatNumber(item.Gift)}</TableCell>
                                    <TableCell className={'flex flex-row]'}>
                                        {item.Guard.split(",").map((item,index) => {
                                            return (
                                                 <div className={'flex flex-row items-center text-primary font-bold w-[109px]'} >
                                                     {
                                                         parseInt(item) !== 0 && (
                                                             <>
                                                                 <img src={getGuardIcon(index+1)}/>
                                                                 <span className={'ml-2'}>{item}</span>
                                                             </>
                                                         )
                                                     }

                                                </div>
                                            )
                                        })}
                                    </TableCell>
                                    <TableCell>{item.Hours} h</TableCell>
                                    <TableCell>{item.Guild}</TableCell>
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </div>
            )}
        </div>
    );
}
const calc = (item) => {
    var init = item.Money
    init = init-item.GuardLiveMoney
    var sum = 0
    var index = 0
    var m = [19998,1998,138]
    item.Guard.split(",").forEach((item) => {
        sum = sum + parseInt(item)*m[index]
        index++
    })

    return sum + init
}
export default ComparePage;
