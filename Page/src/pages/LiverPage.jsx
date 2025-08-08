import React from 'react';
import {useNavigate, useParams} from "react-router-dom";
import axios from "axios";
import {LineChart} from "@mui/x-charts/LineChart";
import {Avatar, Card, CardBody, CardHeader, Image, Tooltip} from "@heroui/react";
import HoverMedals from "../components/HoverMedals";

function LiverPage(props) {
    const [fansChart, setFansChart] = React.useState([]);

    const [guardChart, setGuardChart] = React.useState([]);

    const [space, setSpace] = React.useState({});

    const [lives, setLives] = React.useState([]);




    let {id} = useParams();

    React.useEffect(() => {
        axios.get(`${protocol}://${host}:${port}/api/chart/fans?uid=${id}`).then((response) => {
            setFansChart(response.data.data??[]);
        })
        axios.get(`${protocol}://${host}:${port}/api/liver/space?uid=${id}`).then((response) => {
            setSpace(response.data);
        })
        axios.get(`${protocol}://${host}:${port}/api/chart/guard?uid=${id}`).then((response) => {
            var dst = []
            var map = [19998, 1998, 138]
            response.data.data?.forEach(element => {
                dst.push({
                    UpdatedAt: element.UpdatedAt,
                    Guard: element.Guard.split(",").reduce((a, b) => parseInt(a) + parseInt(b)),
                })
            })
            setGuardChart(dst??[])
        })
        axios.get(`${protocol}://${host}:${port}/api/live?uid=${id}&limit=1000`).then((response) => {
            setLives(response.data.lives);
        })
    }, [])
    return (
        <div>
            <div className={'flex flex-col mb-6 sm:justify-center items-center '}>
                <Avatar src={`${protocol}://${host}:${port}${import.meta.env.PROD ? '' : '/api'}/face?mid=${id}`}
                        size={'lg'} onClick={() => {
                    toSpace(id)
                }}/>
                <div className={'ml-4 flex flex-col items-center'}>
                    <span className={'font-bold'}>{space.UName}</span>
                    {space.Verify &&
                        <div className={'flex '}>
                            <svg style={{color: '#ffcc00'}} width="16" height="16" viewBox="0 0 16 16" fill="none"
                                 xmlns="http://www.w3.org/2000/svg">
                                <path
                                    d="M16 8C16 12.4183 12.4183 16 8 16C3.58172 16 0 12.4183 0 8C0 3.58172 3.58172 0 8 0C12.4183 0 16 3.58172 16 8Z"
                                    fill="currentColor"></path>
                                <path fill-rule="evenodd" clip-rule="evenodd"
                                      d="M7.28832 12.7244C7.20127 12.767 7.1148 12.7988 7.02538 12.8C6.80863 12.8042 6.62919 12.6296 6.62564 12.4101C6.62742 12.3717 6.63512 12.3351 6.64814 12.2997L7.40676 8.78586L4.26392 8.79186C4.03651 8.79545 3.85825 8.6209 3.85352 8.40196C3.85588 8.27299 3.9228 8.15362 4.03118 8.08524L8.72206 3.2901C8.80852 3.23732 8.90149 3.20133 8.99743 3.20013C9.21477 3.19653 9.39303 3.37108 9.39776 3.59063C9.39599 3.65541 9.37822 3.71839 9.34506 3.77358L8.59118 7.23047H11.7388C11.9614 7.22687 12.1403 7.40142 12.1444 7.62096C12.1426 7.75113 12.0757 7.8705 11.9668 7.93888L7.28832 12.7244Z"
                                      fill="white"></path>
                            </svg>
                            <span className={'font-bold ml-2'}>{space.Verify}</span>
                        </div>
                    }

                    <span className={'font-thin text-sm'}>{space.Bio}</span>
                </div>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 text-sm">
                <div
                    className="rounded-xl bg-green-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">粉丝<br/>
                    <span className="font-semibold">{parseInt(space.Fans).toLocaleString()}</span>
                </div>
                <div
                    className="rounded-xl bg-yellow-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">大航海<br/>
                    <span
                        className="font-semibold">{space.Guard}</span>
                </div>
                <Tooltip content={<HoverMedals ruid={id}/>}>
                    <div
                        className="rounded-xl bg-pink-50 p-2 transition-transform duration-200 hover:scale-105 hover:shadow-lg ">粉丝牌<br/>
                        <span
                            className="font-semibold">{space.Medal}</span>
                    </div>
                </Tooltip>
            </div>
            <div className={'grid grid-cols-1 sm:grid-cols-2 w-full'}>
                <LineChart
                    xAxis={[{
                        data: fansChart.map((item, index) => (new Date(item.CreatedAt).getTime())),
                        label: "粉丝",
                        valueFormatter: (timestamp) => {
                            const date = new Date(timestamp)
                            return date.toLocaleDateString()
                        }
                    }]}
                    series={[
                        {
                            data: fansChart.map((item, index) => (item.Fans))
                        },
                    ]}
                    height={300}
                />
                <LineChart
                    xAxis={[{
                        data: guardChart.map((item, index) => (new Date(item.UpdatedAt).getTime())),

                        label: "大航海",
                        valueFormatter: (timestamp) => {
                            const date = new Date(timestamp)
                            return date.toLocaleDateString()
                        }
                    }]}
                    series={[
                        {
                            data: guardChart.map((item, index) => (new Date(item.Guard).getTime())),


                        },
                    ]}
                    height={300}
                />
            </div>
            <div className={'grid grid-cols-1 sm:grid-cols-6'}>
                {lives.map((live, index) => (
                    <LiveStatisticCard item={live}/>
                ))}
            </div>
        </div>
    );
}


function LiveStatisticCard(props) {
    const redirect = useNavigate()
    var item = props.item
    console.log(item)
    return (
        <div onClick={e => {
            redirect('/lives/' + item.ID)
        }}>
            <Card className={'my-4 mx-2'} isHoverable >
                <CardHeader className="flex-col items-start">
                    <p className="text-large uppercase font-bold">{item.Title}</p>
                    <small className="text-default-500">{new Date(item.CreatedAt).toLocaleString()}</small>
                    <div className={'flex-row flex items-center justify-center'}>
                        <svg xmlns="http://www.w3.org/2000/svg" className="icon" viewBox="0 0 1024 1024"
                             style={{width: '20px', height: '20px'}}>
                            <path
                                d="M464 512a48 48 0 1 0 96 0 48 48 0 1 0-96 0zm200 0a48 48 0 1 0 96 0 48 48 0 1 0-96 0zm-400 0a48 48 0 1 0 96 0 48 48 0 1 0-96 0zm661.2-173.6c-22.6-53.7-55-101.9-96.3-143.3a444.35 444.35 0 0 0-143.3-96.3C630.6 75.7 572.2 64 512 64h-2c-60.6.3-119.3 12.3-174.5 35.9a445.35 445.35 0 0 0-142 96.5c-40.9 41.3-73 89.3-95.2 142.8-23 55.4-34.6 114.3-34.3 174.9A449.4 449.4 0 0 0 112 714v152a46 46 0 0 0 46 46h152.1A449.4 449.4 0 0 0 510 960h2.1c59.9 0 118-11.6 172.7-34.3a444.48 444.48 0 0 0 142.8-95.2c41.3-40.9 73.8-88.7 96.5-142 23.6-55.2 35.6-113.9 35.9-174.5.3-60.9-11.5-120-34.8-175.6zm-151.1 438C704 845.8 611 884 512 884h-1.7c-60.3-.3-120.2-15.3-173.1-43.5l-8.4-4.5H188V695.2l-4.5-8.4C155.3 633.9 140.3 574 140 513.7c-.4-99.7 37.7-193.3 107.6-263.8 69.8-70.5 163.1-109.5 262.8-109.9h1.7c50 0 98.5 9.7 144.2 28.9 44.6 18.7 84.6 45.6 119 80 34.3 34.3 61.3 74.4 80 119 19.4 46.2 29.1 95.2 28.9 145.8-.6 99.6-39.7 192.9-110.1 262.7z"/>
                        </svg>
                        <h3 className={'mt-1 ml-0.5'}>{item.Message}</h3>
                        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" style={{width: '20px', height: '20px'}} className={'ml-1'}>
                            <path
                                d="M911.5 700.7a8 8 0 0 0-10.3-4.8L840 718.2V180c0-37.6-30.4-68-68-68H252c-37.6 0-68 30.4-68 68v538.2l-61.3-22.3c-.9-.3-1.8-.5-2.7-.5-4.4 0-8 3.6-8 8V763c0 3.3 2.1 6.3 5.3 7.5L501 910.1c7.1 2.6 14.8 2.6 21.9 0l383.8-139.5c3.2-1.2 5.3-4.2 5.3-7.5v-59.6c0-1-.2-1.9-.5-2.8zM512 837.5l-256-93.1V184h512v560.4l-256 93.1zM660.6 312h-54.5c-3 0-5.8 1.7-7.1 4.4l-84.7 168.8H511l-84.7-168.8a8 8 0 0 0-7.1-4.4h-55.7c-1.3 0-2.6.3-3.8 1-3.9 2.1-5.3 7-3.2 10.8l103.9 191.6h-57c-4.4 0-8 3.6-8 8v27.1c0 4.4 3.6 8 8 8h76v39h-76c-4.4 0-8 3.6-8 8v27.1c0 4.4 3.6 8 8 8h76V704c0 4.4 3.6 8 8 8h49.9c4.4 0 8-3.6 8-8v-63.5h76.3c4.4 0 8-3.6 8-8v-27.1c0-4.4-3.6-8-8-8h-76.3v-39h76.3c4.4 0 8-3.6 8-8v-27.1c0-4.4-3.6-8-8-8H564l103.7-191.6c.6-1.2 1-2.5 1-3.8-.1-4.3-3.7-7.9-8.1-7.9z"/>
                        </svg>
                        <span className=" mt-1 ml-0.5">{item.Money}</span>
                    </div>
                </CardHeader>
                <CardBody className="overflow-visible py-2">
                    <Image
                        alt="Card background"
                        className="object-cover rounded-xl"
                        src={item.Cover}
                        isBlurred
                        isZoomed

                    />
                </CardBody>
            </Card>
        </div>

    )
}

export default LiverPage;