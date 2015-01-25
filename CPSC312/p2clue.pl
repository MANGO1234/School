:- dynamic room/1.
:- dynamic weapon/1.
:- dynamic suspect/1.
:- dynamic inplay/1.
:- dynamic playerTable/2.
:- dynamic playerHas/2.
:- dynamic numCards/2.
:- dynamic numOfPlayer/1.
:- dynamic player/1.
:- dynamic playerOrder/1.
:- dynamic history/4.

% documentation in read-me.txt

/*room(a). room(b). room(c).

weapon(d). weapon(e). weapon(f).

suspect(g). suspect(h). suspect(i).

inplay(a). inplay(b). inplay(c). inplay(d). inplay(e). inplay(f). inplay(g). inplay(h). inplay(i). % quick testing*/


/*suspect(plum). suspect(scarlet). suspect(mustard). suspect(green). suspect(white). suspect(peacock).

room(kitchen). room(ballroom). room(conservatory). room(billiard). room(library). room(study). room(hall). room(lounge). room(dining).

weapon(knife). weapon(candlestick). weapon(revolver). weapon(rope). weapon(pipe). weapon(wrench).

inplay(plum). inplay(scarlet). inplay(mustard). inplay(green). inplay(white). inplay(peacock).
inplay(knife). inplay(candlestick). inplay(revolver). inplay(rope). inplay(pipe). inplay(wrench).
inplay(kitchen). inplay(ballroom). inplay(conservatory). inplay(billiard).
inplay(library). inplay(study). inplay(hall). inplay(lounge). inplay(dining). % testing */

% start a new game
startNewGame :- s.
s :-
	retractall(numOfPlayer(_)),
	retractall(playerOrder(_)),
	retractall(player(_)),
	retractall(playerTable(_, _)),
	retractall(playerHas(_, _)),

	retractall(inplay(_)),
	retractall(room(_)),
	retractall(weapon(_)),
	write('Enter rooms: '), nl, loadRoom,
	write('Enter weapons: '), nl, loadWeapon,
	% preload suspects
	retractall(suspect(_)),
	Suspects = [plum, scarlet, mustard, green, white, peacock],
	forall(member(Suspect, Suspects), assert(suspect(Suspect))),
	forall(member(Suspect, Suspects), assert(inplay(Suspect))),%*/

	write('Enter the number of players: '), 
	readInt(N),
	assert(numOfPlayer(N)),

	write('Enter the players. The first player entered should be you. Then the player to your left, the second player to your left and so on (i.e. clockwise order): '), nl,
	readPlayerOrder([], N, Players),
	assert(playerOrder(Players)),

	write('Enter your cards: '), nl,
	Players = [Player|_],
	readCardsInHand(Player),
	removePlayerFromTable(Player).

loadRoom :-
	write('Enter room (enter \'done\' to stop): '), read(X), (
		X == 'done' -> true;
		true -> assert(room(X)), assert(inplay(X)), loadRoom
	).

loadWeapon :-
	write('Enter weapon (enter \'done\' to stop): '), read(X), (
		X == 'done' -> true;
		true -> assert(weapon(X)), assert(inplay(X)), loadWeapon
	).


readPlayerOrder(XS, 0, YS) :- reverse(XS, YS), !. % cut prevent infinite, reverse to get right order
readPlayerOrder(XS, N, YS) :-
	write('Enter next player: '),
	read(X),
	assert(player(X)),

	write('Enter number of cards the player has: '),
	readInt(Y),
	assert(numCards(X, Y)),

	findall(Z, card(Z), Cards),
	% addPlayerTable(X, Cards),
	forall(member(Card, Cards), assert(playerTable(X, Card))),
	N1 is N - 1,
	readPlayerOrder([X|XS], N1, YS).

addPlayerTable(_, []).
addPlayerTable(X, [Card|Cards]) :- assert(playerTable(X, Card)), addPlayerTable(X, Cards).

readCardsInHand(Player) :-
	write('Enter card (enter \'done\' to stop): '), readCard(Card), (
		Card == 'done' -> true;
		retract(inplay(Card)) -> write('Card successfully recorded.'), knowPlayerHaveCard2(Player, Card), nl, readCardsInHand(Player);
		true -> write('Card not found or already entered.'), nl, readCardsInHand(Player)
	).


% call when you made a suggestion, enter the result of the suggestion
madeSuggestion :- m.
m :-
	write('Enter suggested room: '),
	readRoom(Room),
	write('Enter suggested weapon: '),
	readWeapon(Weapon),
	write('Enter suggested suspect: '),
	readSuspect(Suspect),
	write('Enter the player that disproved your suggestion (enter none if no one disproved it): '),
	readPlayer(Player), (
			% no one disproved it, so check room/weapon/suspect
			Player == 'none' -> noneCard(Room), noneCard(Weapon), noneCard(Suspect);
			true -> write('Enter the card shown: '), readCard(Card), knowPlayerHaveCard2(Player, Card)
		),

	% find players between you and player, (if none, just use yourself as the player)
	% we know that they definitely don't have the suggested cards
	playerOrder(Players),
	[You|_] = Players, (
			Player == 'none' -> getPlayersBetween(You, You, BPlayers);
			true -> getPlayersBetween(You, Player, BPlayers)
		),
	removeCardFromPlayers(Room, BPlayers),
	removeCardFromPlayers(Weapon, BPlayers),
	removeCardFromPlayers(Suspect, BPlayers),

	deductTable,
	checkWin.

% check if we won (well if you can make it to next round for accusation), tell user
checkWin :- win, write('Only one possibility remains: '),
	getRooms(Rooms), Rooms = [Room],
	getWeapons(Weapons), Weapons = [Weapon],
	getSuspects(Suspects), Suspects = [Suspect],
	write(Room), write(', '), write(Weapon), write(', '), write(Suspect), write('.').

% one of each type = win
win :- countRoom(RN), RN == 1, countWeapon(WN), WN == 1, countSuspect(SN), SN == 1.

% idea is if a card is still in play, and nobody has it, it must be in the envelope
noneCard(X) :- inplay(X), knowCardInEnvelope(X).

% print to user we have deduced a card is in envelope, do the necessary removals
knowCardInEnvelope(Card) :- (
	not(inEnvelope(Card)) -> write('Deduced '), write(Card), write(' is in envelope.'), nl, (
		% remove from table, remove everything else from inplay
		room(Card)    -> getRooms(XS),    delete(XS, Card, YS), removeInplays(YS), removeCardFromTable(Card);
		weapon(Card)  -> getWeapons(XS),  delete(XS, Card, YS), removeInplays(YS), removeCardFromTable(Card);
		suspect(Card) -> getSuspects(XS), delete(XS, Card, YS), removeInplays(YS), removeCardFromTable(Card)
		)
	).


deductTable :- findall(X, card(X), Cards), deductCards(Cards),
               playerOrder(Players), deductPlayerHand(Players),
               findall([P,R,W,S], history(P,R,W,S), History), deductHistory(History). 

deductCards([]).
deductCards([Card|Rest]) :- (
	% if no one can have a card, and we don't know where it is, it must be in the envelope, deductTable restart deduction from scratch base on new info
	% inplay(Card), not(inEnvelope(Card)) prevents infinite recursion
	(inplay(Card), not(inEnvelope(Card)), aggregate_all(count, playerTable(_, Card), 0)) -> knowCardInEnvelope(Card), deductTable;
	% in certain situation, there can be entries left in table even though it's the only one in play
	% (e.g. player 5 showed you pipe, and there are only pipe and candle left as weapon, candle must be in envelope)
	(inEnvelope(Card), aggregate_all(count, playerTable(_, Card), C), C > 0)
		-> write('Deduced '), write(Card), write(' is in envelope.'), nl, removeCardFromTable(Card), deductTable;
	true -> deductCards(Rest)
	).

deductPlayerHand([]).
deductPlayerHand([Player|Rest]) :- countPlayerKnownCards(Player, CountK), countPlayerUnknownCards(Player, CountU), TC is CountK + CountU, (
		% If the number of cards a player has in their hand is equal to the number of known cards they have plus the possible cards they might have
		% then they must have those cards.
		numCards(Player, TC), CountU > 0 -> findall(X, playerTable(Player, X), Cards),
			forall(member(Card, Cards), knowPlayerHaveCard(Player, Card)), removePlayerFromTable(Player), deductTable;

		% there are situations where we know what all cards a player have but there are still entries left in the table
		% (e.g. made suggestion, we got the last card a player has)
		numCards(Player, CountK), aggregate_all(count, playerTable(Player, _), C), C > 0 ->
			removePlayerFromTable(Player), deductTable;

		true -> deductPlayerHand(Rest)
	).

countPlayerKnownCards(Player, Count) :- aggregate_all(count, playerHas(Player, _), Count).
countPlayerUnknownCards(Player, Count) :- aggregate_all(count, playerTable(Player, _), Count).

deductHistory([]).
deductHistory([[Player, Room, Weapon, Suspect]|Rest]) :- (
		% this only serves to remove history to speed up (although ther's so little moves this is never a problem)
		% either we know that a player has one of the card so we can't really deduce anything or we know where all 3 cards are
		% so no need to deduce anything either
		(playerHas(Player, room); playerHas(Player, Weapon); playerHas(Player, Suspect)) ->
			retractT(history(Player, Room, Weapon, Suspect)), deductHistory(Rest);
		(playerHas(_, room), playerHas(_, Weapon), playerHas(_, Suspect)) ->
			retractT(history(Player, Room, Weapon, Suspect)), deductHistory(Rest);

		% if we know any 2 cards cannot be in player 2's hand, he must have shown player 1 the remaining card
		(notInPlayerHand(Player, Room), notInPlayerHand(Player, Weapon)) -> 
			retractT(history(Player, Room, Weapon, Suspect)), knowPlayerHaveCard(Player, Suspect), deductTable;
		(notInPlayerHand(Player, Room), notInPlayerHand(Player, Suspect)) -> 
			retractT(history(Player, Room, Weapon, Suspect)), knowPlayerHaveCard(Player, Weapon), deductTable;
		(notInPlayerHand(Player, Weapon), notInPlayerHand(Player, Suspect)) -> 
			retractT(history(Player, Room, Weapon, Suspect)), knowPlayerHaveCard(Player, Room), deductTable;
		deductHistory(Rest)
	).


% a card is in envelope if it's inplay and it's the only one inplay
inEnvelope(Card) :- inplay(Card), (
	room(Card) -> countRoom(Count), Count == 1;
	weapon(Card) -> countWeapon(Count), Count == 1;
	suspect(Card) -> countSuspect(Count), Count == 1).

% a card is not in a player's hand if it's not in playerTable and he doesn't have it
% (which means another player has it or it's in the envelope)
notInPlayerHand(Player, Card) :- not(playerTable(Player, Card)), not(playerHas(Player, Card)).



% call when another player made a suggestion, enter the result
otherSuggestion :- o.
o :-
	write('Which player made a suggestion: '),
	readPlayer(Player1),
	write('Enter suggested room: '),
	readRoom(Room),
	write('Enter suggested weapon: '),
	readWeapon(Weapon),
	write('Enter suggested suspect: '),
	readSuspect(Suspect),
	write('Enter the player that disproved the suggestion (enter none if no one disproved it): '),

	% find players between player 1 and player 2, (if none, just use player 1 in place of player 2)
	% we know that they definitely don't have the suggested cards
	readPlayer(Player2), (
			Player2 = 'none' -> getPlayersBetween(Player1, Player1, Players);
			true -> getPlayersBetween(Player1, Player2, Players)
		),
	removeCardFromPlayers(Room, Players),
	removeCardFromPlayers(Weapon, Players),
	removeCardFromPlayers(Suspect, Players),

	% put the suggestion in history to be used in deductTable
	% only put it in if we don't know where one of the card is
	(
		Player2 \= 'none', (inplay(Room), not(inEnvelope(Room));
			inplay(Weapon), not(inEnvelope(Weapon)); inplay(Suspect), not(inEnvelope(Suspect))) -> assert(history(Player2, Room, Weapon, Suspect));
		true -> true % hmm need this so the remaining operations are called
	),

	% call one more time in case the above changed something
	deductTable,
	checkWin.

% remove a certain card from the given players from the table
removeCardFromPlayers(Card, Players) :- forall(member(Player, Players), removePlayerCardFromTable(Player, Card)).
% removeCardFromPlayers(_, []).
% removeCardFromPlayers(Card, [X|XS]) :- removePlayerCardFromTable(X, Card), removeCardFromPlayers(Card, XS).

% get players between 2 player: duplicate the player order list and append them together
% find P1, then add players to a list until P2 is found and function halt
getPlayersBetween(P1, P2, XS) :-
	playerOrder(Players),
	append(Players, Players, Lookup),
	getPlayersBetweenH(P1, P2, Lookup, XS).

getPlayersBetweenH(P1, P2, [P1|Rest], XS) :- getPlayersBetweenH2(P2, Rest, XS), !.
getPlayersBetweenH(P1, P2, [_|Rest], XS) :- getPlayersBetweenH(P1, P2, Rest, XS).
getPlayersBetweenH2(P2, [P2|_], []) :- !.
getPlayersBetweenH2(P2, [X|Rest], [X|XS]) :- getPlayersBetweenH2(P2, Rest, XS).




% display the information stored in the database
displayNotebook :- d.
d :-
	write('Possible murder room   : '),
	getRooms(Rooms), displayList(Rooms), nl,
	write('Possible murder weapon : '),
	getWeapons(Weapons), displayList(Weapons), nl,
	write('Possible murder suspect: '),
	getSuspects(Suspects), displayList(Suspects), nl,
	nl,
	playerOrder(Players),
	maplist(displayPlayerCard, Players),
	nl,
	maplist(displayPlayerTable, Players).

displayList([]).
displayList([X]) :- write(X), write('.'), !.
displayList([X|XS]) :- write(X), write(', '), displayList(XS).

displayPlayerCard(Player) :-
	getPlayerCards(Player, Cards), length(Cards, L), (
			L \= 0 -> write('Player '), write(Player), write(' definitely have: '), displayList(Cards), nl;
			true -> write('Player '), write(Player), write(' definitely have: unknown'), nl
		).

% may make this prettier if there's time
displayPlayerTable(Player) :-
	getPlayerTableCards(Player, Cards), length(Cards, L), (
			L \= 0 -> write('Player '), write(Player), write(' may have: '), displayList(Cards), nl
		), !.
displayPlayerTable(_).



% ask for a suggestion to make
cheatSuggestion :- c.
c :- win, write('Make the accusation '),
	getRooms(Rooms), Rooms = [Room],
	getWeapons(Weapons), Weapons = [Weapon],
	getSuspects(Suspects), Suspects = [Suspect],
	write(Room), write(', '), write(Weapon), write(', '), write(Suspect), write('.'), !. % cut to prevent rest

% essentially find the cards where we know the least amount about (i.e. a lot of player who can possible have it)
% so we can cross off as many 'cells' in the table as possible in the table when we are shown a card.
% On the other hand if we know the card in the envelope we should choose either a card in our hand with the matching type
% or the card in the envelope so we can get info on the remining cards (none of the other player can have it) so the remaining 2 cards
% will be guaranteed a hit. We will let the player pick which one he wants to use.
c :- write('Suggest one of the following room: '), getSuggestedCard(room, Rooms), displayList(Rooms), nl,
     write('Suggest one of the following weapon: '), getSuggestedCard(weapon, Weapons), displayList(Weapons), nl,
     write('Suggest one of the following suspect: '), getSuggestedCard(suspect, Suspects), displayList(Suspects), nl,
     write('(Tip: given you have to travel to a room to make suggestion, you may choose to use the room you are currently instead of the suggested rooms to maximize efficiency)').

getSuggestedCard(Type, XS) :- countCard(Type, Count), findall(Z, cardT(Type, Z), Cards), (
	Count == 1 -> findPassCard(Type, XS);
	true -> findall(X, maxUnknownCards(Type, Cards, X), XS)
	).

% find a card that none of the oher player can possibly have
findPassCard(Type, XS) :- playerOrder(Players), [You|_] = Players, getPlayerCards(You, Cards), include(Type, Cards, YS1),
	findall(X, inplayT(Type, X), YS2), append(YS1, YS2, XS).

% card is maximally unknown if it has the most possible players that can have it
% therefore we can tick off more cells when we are shown a card
% conveninently, this will never choose cards we already know the location of (because they would have 0 entries in playerTable)
maxUnknownCards(_, [], _).
maxUnknownCards(Type, [X|XS], Card) :- call(Type, Card), possiblePlayer(X, N1), possiblePlayer(Card, N2), N2 >= N1, maxUnknownCards(Type, XS, Card).

possiblePlayer(Card, Count) :- aggregate_all(count, playerTable(_, Card), Count).


% a bunchof function that provides some minimal input checking
readInt(X) :-
	read(Y), (
		integer(Y) -> X = Y;
		true -> write('Please enter an integer: '), readInt(X)
	).

readPlayer(X) :-
	read(Y), (
		player(Y) -> X = Y;
		Y = 'none' -> X = Y;
		true -> write('Please enter a player: '), readPlayer(X)
	).

readSuspect(X) :-
	read(Y), (
		suspect(Y) -> X = Y;
		true -> write('Please enter a suspect: '), readSuspect(X)
	).

readWeapon(X) :-
	read(Y), (
		weapon(Y) -> X = Y;
		true -> write('Please enter a weapon: '), readWeapon(X)
	).


readRoom(X) :-
	read(Y), (
		room(Y) -> X = Y;
		true -> write('Please enter a room: '), readRoom(X)
	).

readCard(X) :-
	read(Y), (
		card(Y) -> X = Y;
		Y = 'done' -> X = Y;
		true -> write('Please enter a card: '), readCard(X)
	).


% a bunch of database manupulation stuff

% testing shows retractall always return true, but retract returns false if it doesn't remove something
% want it to return true because we may be shown same card by same player twice etc.
retractT(T) :- retract(T), !.
retractT(_) :- true.

% also prevent adding redundant facts
assertT(T) :- not(T), assert(T), !.
assertT(_) :- true.


% note we are using retractT, assertT (which always return true, so all clauses always go through).
removeInplay(X) :- retractT(inplay(X)).
removeInplays(XS) :- maplist(removeInplay, XS).

removePlayerCardFromTable(Player, Card) :- retractT(playerTable(Player, Card)).
removePlayerFromTable(X) :- retractall(playerTable(X, _)).
removeCardFromTable(X) :- retractall(playerTable(_, X)).

% when we know a player has a card, remove it from play, remove the possibilit of anyone else having the card,
% and store which player has the card, the 2nd version surpress writing to console informing the user
knowPlayerHaveCard(Player, Card) :- (not(playerHas(Player, Card)) -> write('Deduced '), write(Player), write(' has '), write(Card), nl;true->true),
	removeInplay(Card), removeCardFromTable(Card), assertT(playerHas(Player, Card)).
knowPlayerHaveCard2(Player, Card) :- removeInplay(Card), removeCardFromTable(Card), assertT(playerHas(Player, Card)).


% get a particular kind of cards that are in play
inplayT(T, Card) :- call(T, Card), inplay(Card).

countCard(Type, Count) :- aggregate_all(count, inplayT(Type, _), Count).
countRoom(Count) :- aggregate_all(count, inplayT(room, _), Count).
countWeapon(Count) :- aggregate_all(count, inplayT(weapon, _), Count).
countSuspect(Count) :- aggregate_all(count, inplayT(suspect, _), Count).

getCards(Type, XS) :- findall(X, inplayT(Type, X), XS).
getRooms(XS) :- findall(X, inplayT(room, X), XS).
getWeapons(XS) :- findall(X, inplayT(weapon, X), XS).
getSuspects(XS) :- findall(X, inplayT(suspect, X), XS).


% get all cards we know a player has
getPlayerCards(Player, Cards) :- findall(X, playerHas(Player, X), Cards).

% get all cards we know a player can have
getPlayerTableCards(Player, Cards) :- findall(X, playerTable(Player, X), Cards).

cardT(Type, X) :- call(Type, X).
card(X) :- room(X).
card(X) :- weapon(X).
card(X) :- suspect(X).
