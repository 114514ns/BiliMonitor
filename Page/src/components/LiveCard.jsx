import {Avatar, Badge, Card, CardBody, CardFooter, CardHeader, Image} from "@heroui/react";


function LiveCard({liveData}) {

    const host = location.hostname;

    const port = debug ? 8080 : location.port;

    const protocol = location.protocol.replace(":", "")
    const {Live, UName, UID, Area, Title, Face, Cover} = liveData;
    const cover = `${protocol}://${host}:${port}/proxy?url=${Cover}`


    const toSpace = (id) => {
        window.open("https://space.bilibili.com/" + id)
    }
    return (
        <div>
            <Card style={{
                width: 300,
                marginRight: '20px',
                margin: '15px',
            }}
            >
                <CardHeader className="flex items-center gap-3">
                    {Live != null ? <Badge color={Live ? "success" : "default"} content="">
                        <Avatar src={`${protocol}://${host}:${port}/proxy?url=${Face}`} onClick={() => {toSpace(UID)}}/>
                    </Badge> : <Avatar src={`${protocol}://${host}:${port}/proxy?url=${Face}`} onClick={() => {toSpace(UID)}}/>}

                    <div>
                        <h4 className="font-semibold">{UName}</h4>
                        <p className="text-gray-500">{Title}</p>
                    </div>
                </CardHeader>
                <CardBody style={{overflow: 'hidden'}}>
                    <Image
                        removeWrapper
                        alt="Card background"
                        className="z-0 w-full h-full object-cover"
                        src={cover}
                        isBlurred
                        isZoomed
                    />
                </CardBody>
                <CardFooter className="flex justify-between">
                    {Live == null?<></>:<span className="text-gray-500 font-semibold">Area: {Area}</span>}
                </CardFooter>
            </Card>
        </div>
    )
}

export default LiveCard;